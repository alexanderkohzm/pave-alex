# PAVE Bank Takehome Test - Bill Management

## Overview

This project implements a fees API to manage a bill using encore and temporal.

Requirements:

- Create a new bill
- Able to add line item to an existing open bill
- Able to close an active bill
  - indicate total amount being charged
  - indicate all line item being charged
- Reject line item addition if bill is closed (bill already charged)
- Able to query open/closed bill
- Able to handle different types of currency (GEL, USD)

## Key Features

- Create bills with specified currency (USD, GEL)
- Add line items to bills with automatic currency conversion, `OriginalAmount` in the original currency and `ConvertedAmount` in the converted currency
- Close bills manually or automatically after 30 days
- Query bill based on ID
- Persistent storage of bills and line items to a DB
- Unit testing for `models/money.go`

## Commands

`encore run`

`temporal server start-dev --db-filename your_temporal.db --ui-port 8080`

[Temporalite is being/has been deprecated, so am using a temporal dev server](https://github.com/temporalio/temporalite-archived/issues/202)

`encore db reset {name_of_service}` to reset the db

## Curl Commands

### Create a bill

```
curl -X POST http://localhost:4000/bills \
  -H "Content-Type: application/json" \
  -d '{"currency": "USD"}'

// Expected Results
{
  "ID": "<UNIQUE_IQ>",
  "Currency": "USD",
  "Status": "OPEN",
  "TotalAmount": 0,
  "Items": []
}
```

### Add Line Item

```
curl -X POST http://localhost:4000/bills/:billId/items \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Consultation Fee",
    "amount": 120.50,
    "currency": "USD"
  }'

// expected return
{
  "ID": "<UniqueID>",
  "Currency": "USD",
  "Status": "OPEN",
  "TotalAmount": 0,
  "LineItems": [
    {
      "ID": "<Unique_ID>",
      "BillID": "<Unique_ID>",
      "Description": "Consultation Fee",
      "Amount": 12050, // if currencies are same, return same amount and exchange rate
      "OriginalAmount": 12050,
      "ExchangeRate": "1",
      "Currency": "USD",
      "CreatedAt": "0001-01-01T00:00:00Z"
    },
    {
      "ID": "<UNIQUE_ID>",
      "BillID": "<UNIQUE_ID>",
      "Description": "Consultation Fee",
      "Amount": 4366,
      "OriginalAmount": 12050,
      "ExchangeRate": "0.36231884057971014",
      "Currency": "GEL",
      "CreatedAt": "0001-01-01T00:00:00Z"
    },
  ]
}
```

### Close a Bill

```
curl -X POST http://localhost:4000/bills/:billId/close

// expected return
{
  "ID": "<UNIQUE_ID>",
  "Currency": "USD",
  "Status": "CLOSED",
  "TotalAmount": 12050,
  "LineItems": [
    {
      "ID": "<UNIQUE_ID>",
      "BillID": "<UNIQUE_ID>",
      "Description": "Consultation Fee",
      "Amount": 12050,
      "OriginalAmount": 12050,
      "ExchangeRate": "1",
      "Currency": "USD",
      "CreatedAt": "0001-01-01T00:00:00Z"
    }
  ]
}
```

### Get a Bill

```
curl http://localhost:4000/bills/:billId

// Expected results
{
  "ID": "<UNIQUE ID>",
  "Currency": "USD",
  "Status": "OPEN",
  "TotalAmount": 0,
  "Items": []
}

```

## Database Schema

### Bills Table

| Column       | Type          | Description                                 |
| ------------ | ------------- | ------------------------------------------- |
| id           | VARCHAR(36)   | Primary key, unique identifier for the bill |
| currency     | VARCHAR(3)    | Currency code (USD or GEL)                  |
| status       | VARCHAR(20)   | Bill status (OPEN/CLOSED)                   |
| total_amount | DECIMAL(19,4) | Total amount of the bill                    |
| created_at   | TIMESTAMP     | When the bill was created                   |
| closed_at    | TIMESTAMP     | When the bill was closed (nullable)         |

### Line Items Table

| Column          | Type          | Description                                  |
| --------------- | ------------- | -------------------------------------------- |
| id              | VARCHAR(36)   | Primary key, unique identifier for line item |
| bill_id         | VARCHAR(36)   | Foreign key reference to bills table         |
| description     | TEXT          | Description of the line item                 |
| amount          | BIGINT        | Amount after currency conversion             |
| original_amount | BIGINT        | Original amount before currency conversion   |
| exchange_rate   | DECIMAL(19,4) | Exchange rate used for conversion            |
| currency        | VARCHAR(3)    | Currency code of the line item               |
| created_at      | TIMESTAMP     | When the line item was created               |

## Decisions

### Why Bills fit nicely with a Temporal workflow

A bill is a business process. It is a natural fit for temporal because it is:

- Long Running
- Has changes to state
- Might have dependencies that fail

### Signals vs. Update API

On the surface, signals and updates allow us to mutate the state of our bill. However, the key difference is how Signals are asynchronous and the Update API is synchronous.

Choosing which one to use for adding line items depends on business requirements and expected behaviour with the consumer of the API.

We would use Signals (asynchronous) if:

- We assume line items are always going to be added eventually
- Adding line items might require additional checks (e.g. anti-fraud) that adds latency

We would use the Update API (synchronous) if:

- It is not guaranteed that a line item would be added and the consumer expects immediate feedback (confirmation)

In the case of this tech test, I have implemented it using the Update API because the synchronous nature of waiting for the update and returning the bill and line item seemed more natural for a simple API end point.

The drawback is that the Update API is processed by the single threaded Temporal Worker by spinning up a go routine. This means that we might run into concurrency issues if we do not handle race conditions such as requests for adding line items and closing a bill coming in at the same time.

### Determinism and Idempotency in Temporal

A key feature in Temporal is that our workflows need to deterministic. This is because when a workflow is replayed after a failure, it needs to have the same outcome.

This means that we encapsulate and wrap any non-deterministic code in Temporal Activities. This includes things such as querying a DB and calling an external API. Temporal will save the input and outputs of these Activties after they have been run in the workflow.

The issue with this is that our Activities need to be idempotent. This means that multiple calling of the Activity will lead to the same outcome.

Let's use an example of updating a Bill stored in our DB via an Activity. A non-idempotent update would be adding an amount to the bill. For example:

- Initial bill: 100
- Add amount: 5
- New amount: 105
  Retrying the Temporal Activity might lead to different outcomes because it's dependent on the current state of the bill's amount. If the bill's amount is different (e.g. it is 200), the new amount will be 205.

An idempotent way of approaching it would be to update the entire state. For example:

- Initial bill: 100
- Upsert entire bill with amount: 105
- New Amount: 105
  This means that re-running the Temporal Activity multiple times will result in the same state - a bill with the new amount of 105.

There are drawbacks with Upserts

- You need to handle the entire object (e.g. the full bill) even though only one field changed
- The last-write "wins" and overrides everything else

### Why use "Save" for Bill and Line_Item?

I initially implemented custom functions that matched the business logic for bills and line items. For example

- Create Bill -> CreateBill()
- Close Bill -> CloseBill()
- Upsert Line Item -> Upsert Line Item and Update Bill total amount (Oops! Mixed responsibilities)

However, there were clear drawbacks of doing it this way:

- We could end up having many custom functions with custom SQL queries and logic. This makes maintenance hard!
- Seperating responsibilities could get tricky (like with the original Upsert Line Item activity) if we are not disciplined

Thus, using an Upsert to `save` (similar to how an ORM has a .save() method) seemed intuitive. The drawback is that you need to be careful when updating the fields.

### Using DB to manage state instead of workflow

Initially it seemed easy to just query/update and return the workflow's state.

However, there are several drawbacks

- If state only exists inside workflows, you lose access when they are deleted
- DBs can be indexed and optimised for read and write performance
- Other systems might want to be integrated (e.g. dashboards and reports) and it will need access to the state
- Temporal is meant to orchestrate workflows, not really act as a system of record

### Money

Money and how we handle it in software is an extremely rich (hah!) topic.

At the crux of it are the following:

- How do we represent numbers in our 2-bit binary system?
- What problems occur and what strategies do we employ to handle it?

The principle is **do not use floats** to represent money due to inaccuracies when representing decimals (the famous 0.1 + 0.2 problem).

The strategies include representing money in the smallest denomination with an integer (BigInt) or representing money using Decimals (essentially a string).

Integer Pros

- Technically more performant than strings (but have never used it at scale to see benefits IRL)
- Avoids floating point calculation errors

Integer Cons

- Less "natural" as people don't typically think of money in the smallest denomination
- Need to perform conversions for display in the frontend
- JSON handles numbers using "Number", might lose BigInt precision

Decimals Pros

- "Natural" and easier to comprehend in the database since this reflects how we use moeny in the real world

Decimals Cons

- How do we know how many decimal places to store up to?
- Technically slower than integers

# Appendix

This section includes any notes and questions that I thought of and tried to answer as I was writing this coding test.

It is raw and messy and simply serves as a reference.

### What is Temporal?

Temporal is a workflow orchestration engine. Using Temporal helps us to build durable, stateful, and long-running workflows.

Essentially, temporal handles retries, states, and failures. It keeps track of the inputs and outputs of different tasks (e.g. activities, queries, signals). If a failure occurs, Temporal can replay the workflow from its history and reconstruct the exact state.

### What is Encore?

Encore is a backend framework that helps build and deploy cloud-based APIs faster. It combines infrastructure, code, and deployment into one system.

Like Temporal, it aims to allow developers to focus on writing business logic rather than the scaffolding around it (retries, infra, db...)

### Why we can't use floats to store money

https://stackoverflow.com/questions/3730019/why-not-use-double-or-float-to-represent-currency/3730040#3730040

### Difference between BigInt, Integer, Decimal

Integer/Int - 4 bytes (2.1 billion range) (2^31- 1)
BIGINT - 8 bytes (9.2 quintillion range) (2^63 -1)

Decimal - arbitrary precision numbers (this basically just means numbers can be stored with exact decimal representation)

- Stored as variable-length binary format
- Performance - slower than integers
- When you MUST store fractional values directly

Floats = base2 = can't represent decimals precisely
Think of Decimal as a string of base-10 digits (string-like)

### When first creating a bill, should we save (persist) the bill before or after the workflow is started?

It makes sense to persist the bill initially only AFTER the workflow has started. This is because if the workflow fails to start, then we get inconsistencies as bills are saved in the DB but is not a workflow

With that being said, we should handle the case where if the workflow fails to start, we should return an error to the API consumer.

### Signals and Updates - When do you use them?

Updates are used when you want synchronous messages. You expect the workflow to process the update and return the result. This means that updates are particularly well suited for mutating state, especially when you need confirmation that your state has been mutated. The issue with updates comes in how they are handled by go - they are handled by go routines. This means that you need to gracefully handle race conditions. This can be circumvented by checking the state of the workflow (e.g. if workflow has been closed, you can't update anymore)

Signals, on the other hand, are good for asynchronous messaging. You "fire and forget". This means that signals are good for triggering side effects or events where you don't need acknowledgement of receipts. Signals are handled serially so you don't have to worry about concurrency issues.

### Using a Database with Encore

One of Encore's power is abstracting away infrastructure concerns. In this case, developers don't need to care about provisioning a database (it provisions a local Postgres DB via Docker).

Encore gives each service its own Postgres database. This is part of Encore's goal of treating each service as an isolated unit (i.e. like a microservice).

This means we can

- Define schemas specific to the service
- Encore automatically wires up connections
- Database is private to the service unless you expose it via API

In this case, our `bill` service will have its own isolated Postgres database.

### Idempotency

Idempotency is about predictable, repeatable operations. Calling an idempotent function multiple times with the same input will always produce the same result and side effects as calling it once.

For example, hitting the "STOP" button in a VLC player will always result in the player stopping. "PLAY/PAUSE" is not idempotent because it toggles between the two 'play/pause' states.

### Why upserts in DB?

So the problem with putting DB calls into activities is that... the DB calls themselves are NOT idempotent!

For example, you might be trying to mutate and update the totalAmount of a bill. The state of the bill might change when the workflow was initially initiated versus when the activity is called again in a replay.

So how do we solve this? We can use upserts (update + insert). Upserts update an existing row if a specified value already exists in a table, and insert a new row if the specified value doesn't already exist.

In the case of adding line items in our DB - if the line item's UUID already exists in the DB, we update the line item. If it doesn't exist, we insert it.

https://www.cockroachlabs.com/blog/sql-upsert/
