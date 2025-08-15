
with

rename as (

    select
        id as user_id,
        age,
        status,
        country,
        created_at
    from {{ source('datalake', 'users') }}

)

from rename
