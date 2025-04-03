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
  "ID": "<UNIQUE_IQ>",
  "Currency": "USD",
  "Status": "OPEN",
  "TotalAmount": 0,
  "Items": []
}
```

Add Line Item

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
      "ID": "<UniqueID>",
      "BillID": "<UniqueID>",
      "Description": "Consultation Fee",
      "Amount": 12050, // if currencies are same, return same amount and exchange rate
      "OriginalAmount": 12050,
      "ExchangeRate": "1",
      "Currency": "USD",
      "CreatedAt": "0001-01-01T00:00:00Z"
    }
  ]
}
```

Close a Bill

```
curl -X POST http://localhost:4000/bills/:billId/close
```

Get a Bill

```
curl http://localhost:4000/bills/:billId

// Expected results
{
  "ID": "<UNIQUE ID>",
  "Currency": "USD",
  "Status": "OPEN",
  "TotalAmount": 0,
  "Items": null
}

```
