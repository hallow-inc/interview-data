-- Users
create table users as 
with

statuses as (
    select repeat(['free'], 6) || repeat(['paid'], 2) || repeat(['trial'], 2) as status_array
),

countries as (
    -- If is a typo that needs to be fixed! It should be 'IT' for Italy
    select repeat(['US'], 5) || repeat(['BR'], 3) || ['IF'] || ['FR'] as country_array
),

final as (
    select
        row_number() over () as id,
        (random() * 80 + 18)::int as age,
        status_array[(1 + (random() * (array_length(status_array) - 1))::int)] as status,
        country_array[(1 + (random() * (array_length(country_array) - 1))::int)] as country,
        ('2023-01-01'::DATE + interval (random() * 365 * 24 * 60 * 60) second) as created_at
    from generate_series(1, 10000),
        statuses,
        countries
)

from final;

copy users to '../sample_data/data/users.csv';
------------
-- Sessions
create table sessions as 
with

final as (
    select
        row_number() over () as id,
        (1 + random() * 9999)::int as user_id,
        (1 + random() * 999)::int as content_id,
        random()::int::bool as is_complete,
        ('2024-01-01'::DATE + interval (random() * 365 * 24 * 60 * 60) second) as created_at
    from generate_series(1, 100000)
)

from final;

copy sessions to '../sample_data/data/sessions.csv';

------------
-- Content
create table content as
with 

media_types as (
    select repeat(['video'], 3) || repeat(['audio'], 6) || ['text'] as media_type_array
),

prayer_type as (
    select ['academic'] || ['podcast'] || repeat(['reflection'], 2) || repeat(['lectio_divina'], 2) || repeat(['rosary'], 3) || ['meditation'] as prayer_type_array
),

final as (
    select
        row_number() over () as id,
        media_type_array[(1 + (random() * (array_length(media_type_array) - 1))::int)] as media_type,
        prayer_type_array[(1 + (random() * (array_length(prayer_type_array) - 1))::int)] as prayer_type,
        ('2023-01-01'::DATE + interval (random() * 365 * 24 * 60 * 60) second) as created_at
    from generate_series(1, 1000),
        media_types,
        prayer_type
)

from final;

copy content to '../sample_data/data/content.csv';
