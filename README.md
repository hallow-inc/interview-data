# Hallow Technical Data Interview

This repo is a partial implementation with multiple services that you'll be expected to finish - at least in part. There are still plenty of design decisions to be made and code to be written, so please be prepared to talk through your thought process and why you chose what you did.

### Requirements

- UNIX-based system. No Windows. You _might_ be able to get away with `wsl` but know that hasn't been tested.
- Have Docker Compose ready to run.
- Have a code editor ready to run.
- Be able to share full screen (for both your code and browser for documentation).

### Dos & Don'ts

- Don't use LLMs. While they are certainly a useful tool and used at Hallow, we're more interested in how you think. An LLM would only get in the way of that objective.
- Do use documentation. Seeing how you process information is valuable.
- Do ask questions.
- Do explain your thought process.

### System Architecture

```
       ┌────────────┐
       │Data Server │
 ┌─────│(Mock APIs) │──────┐
 │     └────────────┘      │
 │  ┌───────┐  ┌────────┐  │
 └─▶│ Event │  │ Batch  │◀─┘
    │Handler│  │Pipeline│
    └───────┘  └────────┘
      │  ┌─────────┐  │
      └─▶│ Bucket  │◀─┘
         │---------│
         │Analytics│
         └─────────┘
```


#### Data Server

**No changes will be required by you here**.

This service simply simulates the flow of data in both a push and pull fashion.
Feel free to look at the code to double-check the structure of the payload if you'd like, but those schemas will be provided.


#### Python Batch Pipeline

Here you'll implement a generic batch data pipeline that fetches data from a REST API and uploads it to S3. You'll need to complete two main tasks:

- Implement `fetch_api_data()` - Create a method that fetches data from API endpoints.
- Implement `run_pipeline()` - Build a generic pipeline executor that encapsulates the data flow from API to S3 with support for optional data transformations.

The S3 upload functionality and JSON serialization are already implemented. Two pipeline instances are then created with your implementation to process `users` and `content` entities.

#### Typescript Event Handler

Here you'll implement a webhook-based event processing service that receives events via HTTP and batches them to S3 storage. You'll need to complete two main tasks:

- Implement `/webhook/events` endpoint - Create a POST handler that validates incoming event payloads, filters out `signup` events, and validate events before adding them to the processing batch.
- Complete S3 upload functionality - The `uploadToS3` function framework is provided but needs the S3 key path schema implemented for proper event organization in the data lake.

#### dbt (Analytics)

Here you'll write a dbt model that'll require some analytical SQL.

This task is a little different than the others since we'll want to run the model adhoc. There'll be a service running in the docker compose cluster, prepped and ready to execute the model code. The model can be updated and ran without restarting the cluster.

With the cluster up, you can run the dbt models and check the output with the following commands from root level of the project: 

```bash
docker compose exec dbt_pipeline uv run dbt run
docker compose exec dbt_pipeline duckdb ./results/dbt.duckdb
```

This will open up a database within the container where you can see the data - assuming the `dbt run` was successful.

Use the following to see the tables that were just materialized:

```sql
show tables;
```
