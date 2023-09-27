CREATE TABLE app_permissions (
    id bigint NOT NULL,
    app_id bigint,
    request_method text,
    max_amount bigint,
    budget_renewal text,
    expires_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE app_permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE app_permissions_id_seq OWNED BY app_permissions.id;

CREATE TABLE apps (
    id bigint NOT NULL,
    user_id bigint,
    name text,
    description text,
    nostr_pubkey text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE apps_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE apps_id_seq OWNED BY apps.id;

CREATE TABLE identities (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    privkey text
);

CREATE SEQUENCE identities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE identities_id_seq OWNED BY identities.id;

CREATE TABLE nostr_events (
    id bigint NOT NULL,
    app_id bigint,
    nostr_id text,
    reply_id text,
    content text,
    state text,
    replied_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE nostr_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE nostr_events_id_seq OWNED BY nostr_events.id;

CREATE TABLE payments (
    id bigint NOT NULL,
    app_id bigint,
    nostr_event_id bigint,
    amount bigint,
    payment_request text,
    preimage text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE payments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE payments_id_seq OWNED BY payments.id;

CREATE TABLE users (
    id bigint NOT NULL,
    alby_identifier text,
    access_token text,
    refresh_token text,
    email text,
    expiry timestamp with time zone,
    lightning_address text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE users_id_seq OWNED BY users.id;

ALTER TABLE ONLY app_permissions ALTER COLUMN id SET DEFAULT nextval('app_permissions_id_seq'::regclass);

ALTER TABLE ONLY apps ALTER COLUMN id SET DEFAULT nextval('apps_id_seq'::regclass);

ALTER TABLE ONLY identities ALTER COLUMN id SET DEFAULT nextval('identities_id_seq'::regclass);

ALTER TABLE ONLY nostr_events ALTER COLUMN id SET DEFAULT nextval('nostr_events_id_seq'::regclass);

ALTER TABLE ONLY payments ALTER COLUMN id SET DEFAULT nextval('payments_id_seq'::regclass);

ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);

ALTER TABLE ONLY app_permissions
    ADD CONSTRAINT app_permissions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY apps
    ADD CONSTRAINT apps_pkey PRIMARY KEY (id);

ALTER TABLE ONLY identities
    ADD CONSTRAINT identities_pkey PRIMARY KEY (id);

ALTER TABLE ONLY nostr_events
    ADD CONSTRAINT nostr_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE INDEX idx_app_permissions_app_id ON app_permissions USING btree (app_id);

CREATE INDEX idx_app_permissions_request_method ON app_permissions USING btree (request_method);

CREATE INDEX idx_apps_nostr_pubkey ON apps USING btree (nostr_pubkey);

CREATE INDEX idx_apps_user_id ON apps USING btree (user_id);

CREATE INDEX idx_identities_deleted_at ON identities USING btree (deleted_at);

CREATE INDEX idx_nostr_events_app_id ON nostr_events USING btree (app_id);

CREATE UNIQUE INDEX idx_nostr_events_nostr_id ON nostr_events USING btree (nostr_id);

CREATE INDEX idx_payments_app_id ON payments USING btree (app_id);

CREATE INDEX idx_payments_nostr_event_id ON payments USING btree (nostr_event_id);

CREATE UNIQUE INDEX idx_users_alby_identifier ON users USING btree (alby_identifier);

ALTER TABLE ONLY app_permissions
    ADD CONSTRAINT fk_app_permissions_app FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE;

ALTER TABLE ONLY nostr_events
    ADD CONSTRAINT fk_nostr_events_app FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE;

ALTER TABLE ONLY payments
    ADD CONSTRAINT fk_payments_app FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE;

ALTER TABLE ONLY payments
    ADD CONSTRAINT fk_payments_nostr_event FOREIGN KEY (nostr_event_id) REFERENCES nostr_events(id);

ALTER TABLE ONLY apps
    ADD CONSTRAINT fk_users_apps FOREIGN KEY (user_id) REFERENCES users(id);