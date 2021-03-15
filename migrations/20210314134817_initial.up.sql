create table feeds (
    id varchar(40) primary key ,
    feed_url text,
    feed_data text
);
create unique index feeds_feeds_url_idx on feeds(feed_url);

create table items (
    id varchar(40) primary key ,
    source_id varchar(40) references feeds(id),
    item_data text
);

create table categories (
    id varchar(40) primary key,
    category_name text
);
create unique index categories_category_name_idx on categories(category_name);

create table item_categories (
    id BIGSERIAL primary key,
    item_id varchar(40) references items(id),
    category_id varchar(40) references categories(id)
);
create unique index item_categories_idx on item_categories(item_id, category_id);

