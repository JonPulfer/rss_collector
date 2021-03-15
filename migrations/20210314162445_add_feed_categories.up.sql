create table feed_categories (
    id BIGSERIAL primary key,
    feed_id varchar(40) references feeds(id),
    category_id varchar(40) references categories(id)
);
create index feed_categories_idx on feed_categories(feed_id, category_id);
