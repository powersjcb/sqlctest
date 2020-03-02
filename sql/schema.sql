CREATE TABLE authors (
                         id BIGSERIAL PRIMARY KEY,
                         name varchar(1024)
);

CREATE TABLE books (
                       id BIGSERIAL PRIMARY KEY,
                       title varchar(4096),
                       author_id bigint,
                       isbn varchar(13),
                       FOREIGN KEY (author_id) REFERENCES authors (id)
);