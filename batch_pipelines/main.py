import json
import os
import time
from datetime import datetime
from typing import Any, Callable, Optional

import boto3
import httpx
from botocore.client import Config

DATA_BASE_URL = os.getenv("DATA_BASE_URL") or "http://localhost:9090/api/"
S3_BUCKET = "datalake"
S3_ENDPOINT = os.getenv("S3_ENDPOINT") or "http://localhost:9000"
S3_ROOT_USER = "adminuser"
S3_ROOT_PASSWORD = "admin123"


class DataPipeline:
    def __init__(self, entity: str):
        self.entity = entity
        self.s3 = boto3.client(
            "s3",
            endpoint_url=S3_ENDPOINT,
            aws_access_key_id=S3_ROOT_USER,
            aws_secret_access_key=S3_ROOT_PASSWORD,
            region_name="us-east-1",  # Ignored but required
            config=Config(signature_version="s3v4"),
            use_ssl=False,
        )

    # TODO: Implement fetching data from the data server.
    # This method should be generic and able to handle different endpoints - in this case `GET /users` and `GET /content`.
    # The method that uploads to S3 is handled for you, so you don't have to worry about the S3 api or client.
    # NOTE: The data will be serialized to JSON before being uploaded to S3.
    # Don't worry about handling that yourself!
    #
    # Tips:
    #
    # - Use the `httpx` library to make GET requests to the data server API endpoints.
    #    i.e., response = httpx.get(DATA_BASE_URL + self.entity, timeout=10, params={"key": "value"})
    # - The API endpoints return JSON data, so you can use `response.json()` to parse the response. To raise an error should one exist, you can use `response.raise_for_status()`.
    #
    #
    # API Endpoints:
    #   - `GET /users` - Fetches user data in a paginated fashion. The max limit is 200 per request.
    #   - `GET /content` - Fetches content data in whole.
    #
    # The payloads are as follows:
    #
    # `users` payload schema:
    #
    # {
    #   count: int,
    #   has_more: true|false,
    #   limit: 200,
    #   next_offset: int,
    #   offset: int,
    #   total_users: 10000,
    #   users: [
    #     {
    #       user_id: str, # i.e, "user_1",
    #       age: int,
    #       status: "free|paid|trial",
    #       country: "US|BR|IT|FR",
    #       created_at: str # i.e., "2023-11-08T00:00:00Z"
    #     },
    #     ...
    #   ]
    # }
    #
    # `content` payload schema:
    #
    # {
    #   count: int,
    #   total_content: 1000,
    #   content: [
    #      {
    #        content_id: str,
    #        prayer_type: "academic|podcast|reflection|lectio_divina|rosary|meditation",
    #        media_type: "audio|video|text",
    #        created_at: str # i.e., "2023-11-08T00:00:00Z"
    #      },
    #      ...
    #   ]
    # }
    #
    def fetch_api_data(self) -> list[dict[str, Any]]:
        """Fetch data from the data server API endpoints"""
        # TODO: Add error handling and retries

        return [{}]  # Replace with actual data from response

    def upload_to_s3(self, data, key: str):
        """Upload data to the S3 bucket"""
        json_data = json.dumps(data, indent=4)
        self.s3.put_object(
            Bucket=S3_BUCKET,
            Key=key,
            Body=json_data.encode("utf-8"),
            ContentType="application/json",
        )

    # TODO: Write a generic pipeline executor that can be used to run any basic pipeline.
    # The `func` parameter is provided to allow for custom transformations and/or metadata injection.
    # This will need to work for both the `users` and `content` pipelines.
    #
    # Tasks:
    #   - Use the fetch_api_data method to get data from the API
    #   - Set the S3 key path so it's appropriately organized within the data lake (S3 bucket)
    #   - Send the data to S3
    #
    def run_pipeline(self, func: Optional[Callable] = None) -> None:
        """Pipeline executor"""


def main():
    time.sleep(20)

    users_pipeline = DataPipeline(entity="users")
    # NOTE: If you have a function for the users pipeline, it'll need to be passed in below.
    users_pipeline.run_pipeline()
    print(
        f"🚰 `users` batch pipelines completed at {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"
    )

    content_pipeline = DataPipeline(entity="content")
    # NOTE: If you have a function for the content pipeline, it'll need to be passed in below.
    content_pipeline.run_pipeline()
    print(
        f"🚰 `content` batch pipelines completed at {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"
    )


if __name__ == "__main__":
    main()
