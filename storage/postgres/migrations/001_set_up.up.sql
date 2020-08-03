create table events
(
    id       varchar(50) not null primary key,
    group_id varchar(50) not null,
    data     bytea       not null
)