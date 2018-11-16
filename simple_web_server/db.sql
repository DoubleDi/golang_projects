
create database if not exists mailru_test;
use mailru_test;
create table if not exists users (id integer primary key, age tinyint, sex enum("M","F"));
create table if not exists stats (user_id integer, action enum("like","comment","exit","login"), ts datetime); 
create index action_ts_user_id_idx on stats (action, ts, user_id);    