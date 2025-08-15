
with

rename as (

    select
        id as session_id,
        user_id,
        content_id,
        is_complete,
        created_at,
    from {{ source('datalake', 'sessions') }}

)

from rename
