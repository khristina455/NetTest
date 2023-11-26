drop table if exists user CASCADE;
drop table if exists modeling CASCADE;
drop table if exists analysis_request CASCADE;
drop table if exists analysis_request_modeling CASCADE;

CREATE TABLE user
(
    user_id           serial          PRIMARY KEY,
    login             varchar(40)     NOT NULL ,
    is_admin          boolean         DEFAULT FALSE,
    name              varchar(40)     NOT NULL,
    password          varchar(64)     NOT NULL,
);
--таблица услуг
CREATE TABLE modeling
(
    modeling_id       serial          PRIMARY KEY,
    name              varchar(60)     NOT NULL,
    description       text            NOT NULL,
    image             text            NOT NULL,
    is_deleted        boolean         NOT NULL,
    price              int            NOT NULL,
);
--таблица заявок
create table analysis_request
(
    request_id     SERIAL unique           not null
        constraint monitoring_requests_pk
            primary key,
    creator_id     int
        constraint creator_id_fk
            references "users" (user_id),
    status         varchar(20)             not null,
    creation_date  timestamp default now() not null,
    formation_date timestamp,
    complete_date    timestamp,
    admin_id       int
        constraint monitoring_request_user_id_fk
            references "users" (user_id)
);
-- таблица связи м:м
create table analysis_request_modeling
(
    id         serial not null,
    request_id int
        constraint request_id_fk
            references monitoring_requests(request_id),
    modeling_id  int
        constraint modeling_id_fk
            references modeling(modeling_id)
);
