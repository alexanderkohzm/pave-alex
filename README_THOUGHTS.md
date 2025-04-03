## Money Money Money... it's a rich man's world

### Definitely can't use floats

https://stackoverflow.com/questions/3730019/why-not-use-double-or-float-to-represent-currency/3730040#3730040

### So... BigInt? Integer? Decimal?

Integer/Int - 4 bytes (2.1 billion range) (2^31- 1)
BIGINT - 8 bytes (9.2 quintillion range) (2^63 -1)

Decimal - arbitrary precision numbers (this basically just means numbers can be stored with exact decimal representation)

- Stored as variable-length binary format
- Performance - slower than integers
- When you MUST store fractional values directly

Floats = base2 = can't represent decimals precisely
Think of Decimal as a string of base-10 digits (string-like)

## Workflow and Lifecycle of Bill

### When first creating a bill, should we save (persist) the bill before or after the workflow is started?

It makes sense to persist the bill initially only AFTER the workflow has started. This is because if the workflow fails to start, then we get inconsistencies as bills are saved in the DB but is not a workflow

With that being said, we should handle the case where if the workflow fails to start, we should return an error to the API consumer.

## Updates vs. Signals

### What are they?

Updates and Signals are two ways to communicate with a Temporal workflow from the "outside" world. For example, a HTTP API call hitting your encore server.

### When do you use them?

Updates are used when you want synchronous messages. You expect the workflow to process the update and return the result. This means that updates are particularly well suited for mutating state, especially when you need confirmation that your state has been mutated. The issue with updates comes in how they are handled by go - they are handled by go routines. This means that you need to gracefully handle race conditions. This can be circumvented by checking the state of the workflow (e.g. if workflow has been closed, you can't update anymore)

Signals, on the other hand, are good for asynchronous messaging. You "fire and forget". This means that signals are good for triggering side effects or events where you don't need acknowledgement of receipts. Signals are handled serially so you don't have to worry about concurrency issues.

### Why should we use either of them for our workflow?

I have selected Updates for our close bill and add line item workflow.

The core reason is that

- I want our API calls (Add Item, Close) to explicitly return the mutated state.
- Reduce the number of calls to either the DB or queries to the workflow to return the updated state

The drawback is that I need to handle the concurrency issue - race conditions such as when there is an attempt to add a line item to an already closed bill.

## Activities

### Using a Database with Encore

One of Encore's power is abstracting away infrastructure concerns. In this case, developers don't need to care about provisioning a database (it provisions a local Postgres DB via Docker).

Encore gives each service its own Postgres database. This is part of Encore's goal of treating each service as an isolated unit (i.e. like a microservice).

This means we can

- Define schemas specific to the service
- Encore automatically wires up connections
- Database is private to the service unless you expose it via API

In this case, our `bill` service will have its own isolated Postgres database.

## Idempotency

Idempotency is about predictable, repeatable operations. Calling an idempotent function multiple times with the same input will always produce the same result and side effects as calling it once.

For example, hitting the "STOP" button in a VLC player will always result in the player stopping. "PLAY/PAUSE" is not idempotent because it toggles between the two 'play/pause' states.

### Why is Idempotency important in Bills?

Close bill should be idempotent. No matter how many calls are made, the same effect of closing the bill should be achieved.

### Idempotency and Temporal Activities

Temporal activities and idempotency go hand in hand. Temporal may retry activities (e.g. on failure, timeout, replay).

Non-idempotent activities can result in inconsistent or duplicate side effects unless we guard against it.

### Why do we put DB calls in activities and not workflow code?

Temporal workflows are deterministic. This means that they MUST produce the same result every time they're replaced (e.g. when Temporal wants to perform a crash recovery).

The issue is that DB calls are non-deterministic. The database state might change between replays.

By putting non-deterministic calls in activities, we are able to produce the same results in replays. This is because if the workflow crashes and replays, it will not re-run the acitvity. Instead, it replays from the result.

### Why upserts in DB?

So the problem with putting DB calls into activities is that... the DB calls themselves are NOT idempotent!

For example, you might be trying to mutate and update the totalAmount of a bill. The state of the bill might change when the workflow was initially initiated versus when the activity is called again in a replay.

So how do we solve this? We can use upserts (update + insert). Upserts update an existing row if a specified value already exists in a table, and insert a new row if the specified value doesn't already exist.

In the case of adding line items in our DB - if the line item's UUID already exists in the DB, we update the line item. If it doesn't exist, we insert it.

https://www.cockroachlabs.com/blog/sql-upsert/

### Woah woah woah, but how does upserts ensure idempotency? What if the line item is updated with different values?

This is a good question. I think this falls back to our definition of what idempotency means in this context.

(A) If idempotency is the intention that "no matter what happens, I just want my LineItem to be created or updated", then yes, upserting ensures this expectations.

(B) However, if idempotency is the intention that "I want my LineItem to be created or updated with the exact values I provide", then no, upserts do not ensure this expectation.

The key here is:

- Is idempotency not wanting to have duplicates? (A)
- Is idempotency wanting the exact same content? (B)

### What are some strategies for idempotency keys?

NOTE: Our idempotency key is our UUID. We are making the assumption that each call from the frontend is unique and valid (i.e., not duplicating). There are strategies to prevent the frontend from sending repeated requests (e.g. a middleware layer to check and see if the frontend has sent an idempotency key we have seen before)

So it really depends on what the intent is and how it fits business needs.

1. Always unique - UUIDv4()
   For example, if we have a guarantee that all entities are unique, then we can just use UUIDv4() as the idempotency key. This means that we will always retry and create the entity.

2. Deterministic Composite ID
   In my lab data extraction, we use a composite of `<lab_id>_<test_id>_<timestamp>` where timestamp is the time of the test.

We made the assumption that if the test (lab+test IDs) has the same timestamp, it is likely the same test as a patient wouldn't really have multiple tests done at the same time.

The trade off is that this might not work in cases where a patient has done multiple tests or if multiple unique tests have the same timestamp as they were generated by the medical provider at the same time (even though there were done at different times).

3. Using the content of the entity as the idempotency key
   Perform a hashing of the content of the entity and use that as the idempotency key. This is good if there are a lot of identical identifiers (e.g. id, sku, order_id) but there are discrepancies in the data.

Trade off is that small differences in content that is semantically the same but structurally different will be considered different. For example, "burger " and "burger" are considered different even though there is just a space.

https://zelark.github.io/nano-id-cc/
