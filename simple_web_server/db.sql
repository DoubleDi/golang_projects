create database if not exists mailru_test;
use mailru_test;
create table if not exists users (id integer primary key, age tinyint, sex enum("M","F"));
create table if not exists stats (user_id integer, action varchar(40), ts datetime); 