CREATE DOMAIN uint8 AS smallint
   CHECK(VALUE >= 0 AND VALUE < 256);

CREATE TABLE IF NOT EXISTS Memo (
    id bigserial primary key,
    mtype uint8 not null,
    mlength uint8
);

CREATE TABLE IF NOT EXISTS Report (
    id bigserial primary key,
    rvk bytea not null,
    tck_bytes bytea not null,
    j_1 uint8 not null,
    j_2 uint8 not null,
    memo_id bigserial references Memo(id),
    timestamp timestamp default current_timestamp
);