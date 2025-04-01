`temporalite start --headless=false`

`temporal server start-dev --db-filename your_temporal.db --ui-port 8080`

`encore run`

# Curl Commands

Create a bill

```
curl -X POST http://localhost:4000/bills \
  -H "Content-Type: application/json" \
  -d '{"currency": "USD"}'

// Expected Results
{
  "id": "<UNIQUE_IQ>",
  "currency": "USD",
  "status": "OPEN",
  "items": []
}
```

Add Line Item

```
curl -X POST http://localhost:4000/bills/:billId/items \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Consultation Fee",
    "amount": 120.50
  }'
```

Close a Bill

```
curl -X POST http://localhost:4000/bills/:billId/close
```

Get a Bill

```
curl http://localhost:4000/bills/:billId
```
