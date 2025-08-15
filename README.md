# Hallow Sr. Data Interview

This repo is a partial implementation with multiple systems that you'll be expected to finish - at least in part. There are still plenty of design decisions to be made, so please be prepared to talk through your thought process and why you chose what you did.

### Requirements

- UNIX-based system. No Windows. You _might_ be able to get away with `wsl` but know that hasn't been tested.
- You can use any tools/LLMs you wish, but you must be explicit about it and show _how_ you're using it. Any shadow use will immediately end the interview.


### Architecture

The system 

┌─────────────────┐    ┌──────────────────┐
│   Data Server   │───▶│  Python Pipeline │─────────────┐
│   (Mock APIs)   │    │   (Batch)        │             │
└─────────────────┘    └──────────────────┘             │
           │                                            ▼
           │           ┌──────────────────┐    ┌──────────────────┐
           └──────────▶│  Typescript      │───▶│       dbt        │
                       │   Event Handler  │    │   (Analytics)    │
                       └──────────────────┘    └──────────────────┘

#### Data Server

**No changes will be required by you here**.

This service simply simulates the flow of data in both a push and pull fashion.


#### Python Batch Pipeline




#### Typescript Event Handler



#### dbt (Analytics)

A simple model set that'll require some data cleaning and sql aggregations.

You will need to go into `./dbt_pipeline/models/mart/__sources.yml` and set the correct key paths based on what you set in the batch pipeline and event handler.

After starting the cluster and allowing for the pipelines to run, run the dbt models with the following commands:

```bash
docker-compose exec your-service uv run dbt run
docker-compose exec your-service uv run dbt test
```
