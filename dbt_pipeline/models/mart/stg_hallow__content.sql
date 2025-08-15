
with

rename as (

    select
        id as content_id,
        media_type,
        prayer_type,
        created_at
    from {{ source('datalake', 'content') }}

)

from rename
