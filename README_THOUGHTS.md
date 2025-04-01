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
