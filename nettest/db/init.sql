drop table if exists "users" CASCADE;
drop table if exists modelings CASCADE;
drop table if exists analysis_requests CASCADE;
drop table if exists analysis_requests_modelings CASCADE;

CREATE TABLE "users"
(
    user_id           serial          PRIMARY KEY,
    login             varchar(40)     NOT NULL ,
    is_admin          boolean         DEFAULT FALSE,
    name              varchar(40)     NOT NULL,
    password          varchar(64)     NOT NULL
);
--таблица услуг
CREATE TABLE modelings
(
    modeling_id       serial          PRIMARY KEY,
    name              varchar(60)     NOT NULL,
    description       text            NOT NULL,
    image             text            ,
    is_deleted        boolean         NOT NULL,
    price             decimal(30, 2)  NOT NULL
);
--таблица заявок
create table analysis_requests
(
    request_id         serial          PRIMARY KEY,
    user_id            int,
    status             varchar(20)     CHECK (status IN ('DRAFT', 'REGISTERED', 'IN WORK', 'COMPLETE', 'CANCELED', 'DELETED')) DEFAULT 'DRAFT',
    creation_date      timestamp       default now() not null,
    formation_date     timestamp,
    complete_date      timestamp,
    admin_id           int             DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES "users"(user_id),
    FOREIGN KEY (admin_id) REFERENCES "users"(user_id)
);
-- таблица связи м:м
create table analysis_requests_modelings
(
    request_id         int,
    modeling_id        int,
    node_quantity      int              NOT NULL,
    queue_size         int              NOT NULL,
    client_quantity    int              NOT NULL,
    result             decimal(30, 2),
    PRIMARY KEY (modeling_id, request_id),
    FOREIGN KEY (request_id)     REFERENCES analysis_requests(request_id),
    FOREIGN KEY (modeling_id)    REFERENCES modelings(modeling_id)
);

INSERT INTO modelings(name, description, image, is_deleted, price)
VALUES ('Аналитическое моделирование очереди в узле сети', 'Рассчет времени ожидания в узле сети',
        '/images/card1.jpg', false, 1500),
       ('Аналитическое моедлирование прохождения сообщения в сети', 'Рассчет среденего времени прохождения сообщения в сети',
        '/images/card2.jpeg', false, 4000);

