CREATE TABLE hops.ai (
	id serial NOT NULL,
	pos_id int4 NOT NULL,
	"server" int4 NOT NULL,
	faction int4 NOT NULL,
	"map" int2 NOT NULL,
	coordinate point NOT NULL,
	walking_speed float4 NOT NULL,
	running_speed float4 NOT NULL,
	CONSTRAINT ai_pkey PRIMARY KEY (id)
);
ALTER TABLE hops.ai ADD CONSTRAINT ai_pos_id_fkey FOREIGN KEY (pos_id) REFERENCES data.npc_pos_table(id);

CREATE TABLE hops.cash (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	amount int4 NOT NULL DEFAULT 0,
	CONSTRAINT cash_pkey PRIMARY KEY (id)
);

CREATE TABLE hops."characters" (
	id serial NOT NULL,
	user_id uuid NOT NULL,
	"name" varchar(32) NOT NULL,
	epoch int8 NOT NULL,
	"type" int4 NOT NULL,
	faction int4 NOT NULL,
	height int4 NOT NULL,
	"level" int4 NOT NULL,
	"class" int4 NOT NULL,
	is_online bool NOT NULL,
	is_active bool NOT NULL,
	gold int8 NOT NULL,
	coordinate point NOT NULL,
	"map" int2 NOT NULL,
	"exp" int8 NOT NULL,
	ht_visibility int4 NOT NULL,
	weapon_slot int4 NOT NULL,
	running_speed float4 NOT NULL,
	guild_id int4 NOT NULL,
	exp_multiplier float4 NOT NULL,
	drop_multiplier float4 NOT NULL,
	slotbar bytea NOT NULL,
	created_at timestamptz NOT NULL,
	additional_exp_multiplier float4 NOT NULL DEFAULT 0,
	additional_drop_multiplier float4 NOT NULL DEFAULT 0,
	aid_mode bool NOT NULL DEFAULT false,
	aid_time int4 NOT NULL DEFAULT 7200,
	CONSTRAINT characters_pkey PRIMARY KEY (id)
);

CREATE TABLE hops.characters_buffs (
	id int4 NOT NULL,
	character_id int4 NOT NULL,
	"name" text NOT NULL,
	atk int4 NOT NULL,
	atk_rate int4 NOT NULL,
	arts_atk int4 NOT NULL,
	arts_atk_rate int4 NOT NULL,
	poison_def int4 NOT NULL,
	paralysis_def int4 NOT NULL,
	confusion_def int4 NOT NULL,
	def int4 NOT NULL,
	def_rate int4 NOT NULL,
	arts_def int4 NOT NULL,
	arts_def_rate int4 NOT NULL,
	accuracy int4 NOT NULL,
	dodge int4 NOT NULL,
	max_hp int4 NOT NULL,
	hp_recovery_rate int4 NOT NULL,
	max_chi int4 NOT NULL,
	chi_recovery_rate int4 NOT NULL,
	str int4 NOT NULL,
	dex int4 NOT NULL,
	"int" int4 NOT NULL,
	exp_multiplier int4 NOT NULL,
	drop_multiplier int4 NOT NULL,
	running_speed float4 NOT NULL,
	started_at int8 NOT NULL,
	duration int8 NOT NULL,
	bag_expansion bool NULL DEFAULT false,
	CONSTRAINT characters_buffs_pkey PRIMARY KEY (id, character_id)
);

CREATE TABLE hops.consignment (
	id int4 NOT NULL,
	seller_id int4 NOT NULL,
	item_name text NOT NULL,
	quantity int4 NOT NULL,
	price int8 NOT NULL,
	is_sold bool NOT NULL DEFAULT false,
	expires_at timestamptz NOT NULL,
	CONSTRAINT consignment_pkey PRIMARY KEY (id)
);

CREATE TABLE hops.guilds (
	id serial NOT NULL,
	announcement text NOT NULL DEFAULT ''::text,
	description text NOT NULL DEFAULT ''::text,
	faction int4 NOT NULL,
	gold_donation int8 NOT NULL DEFAULT 0,
	honor_donation int8 NOT NULL DEFAULT 0,
	logo bytea NOT NULL,
	member_count int4 NOT NULL DEFAULT 0,
	members jsonb NOT NULL DEFAULT '{}'::jsonb,
	"name" text NOT NULL,
	recognition int8 NOT NULL DEFAULT 0,
	leader_id int4 NOT NULL,
	CONSTRAINT guilds_pkey PRIMARY KEY (id)
);

CREATE TABLE hops.items_characters (
	id serial NOT NULL,
	user_id text NULL,
	character_id int4 NULL,
	item_id int4 NOT NULL,
	slot_id int4 NOT NULL,
	quantity int4 NOT NULL DEFAULT 0,
	plus int4 NOT NULL DEFAULT 0,
	upgrades _int4 NOT NULL DEFAULT '{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}'::integer[],
	socket_count int4 NOT NULL DEFAULT 0,
	sockets _int4 NOT NULL DEFAULT '{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}'::integer[],
	activated bool NOT NULL DEFAULT false,
	in_use bool NOT NULL DEFAULT false,
	pet_info jsonb NULL,
	updated_at timestamptz NULL,
	consignment bool NOT NULL DEFAULT false,
	CONSTRAINT items_characters_pkey PRIMARY KEY (id)
);

CREATE TABLE hops.relics (
	id int4 NOT NULL,
	count int4 NULL DEFAULT 0,
	"limit" int4 NOT NULL,
	tradable bool NULL DEFAULT false,
	required_items _int4 NULL,
	CONSTRAINT relics_pkey PRIMARY KEY (id)
);

CREATE TABLE hops.servers (
	id int4 NOT NULL,
	"name" varchar(32) NOT NULL,
	max_users int4 NOT NULL
);

CREATE TABLE hops.skills (
	id int4 NOT NULL,
	skill_points int4 NOT NULL,
	skills jsonb NOT NULL DEFAULT '{"slots": [{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}]}'::jsonb,
	CONSTRAINT skills_pkey PRIMARY KEY (id)
);

CREATE TABLE hops.stats (
	id int4 NOT NULL,
	hp int4 NOT NULL DEFAULT 0,
	max_hp int4 NOT NULL DEFAULT 0,
	hp_recovery_rate int4 NOT NULL DEFAULT 0,
	chi int4 NOT NULL DEFAULT 0,
	max_chi int4 NOT NULL DEFAULT 0,
	chi_recovery_rate int4 NOT NULL DEFAULT 0,
	str int4 NOT NULL DEFAULT 0,
	dex int4 NOT NULL DEFAULT 0,
	"int" int4 NOT NULL DEFAULT 0,
	str_buff int4 NOT NULL DEFAULT 0,
	dex_buff int4 NOT NULL DEFAULT 0,
	int_buff int4 NOT NULL DEFAULT 0,
	stat_points int4 NOT NULL DEFAULT 0,
	honor int4 NOT NULL DEFAULT 0,
	min_atk int4 NOT NULL DEFAULT 0,
	max_atk int4 NOT NULL DEFAULT 0,
	atk_rate int4 NOT NULL DEFAULT 0,
	min_arts_atk int4 NOT NULL DEFAULT 0,
	max_arts_atk int4 NOT NULL DEFAULT 0,
	arts_atk_rate int4 NOT NULL DEFAULT 0,
	def int4 NOT NULL DEFAULT 0,
	def_rate int4 NOT NULL DEFAULT 0,
	arts_def int4 NOT NULL DEFAULT 0,
	arts_def_rate int4 NOT NULL DEFAULT 0,
	accuracy int4 NOT NULL DEFAULT 0,
	dodge int4 NOT NULL DEFAULT 0,
	poison_atk int4 NOT NULL DEFAULT 0,
	paralysis_atk int4 NOT NULL DEFAULT 0,
	confusion_atk int4 NOT NULL DEFAULT 0,
	poison_def int4 NOT NULL DEFAULT 0,
	paralysis_def int4 NOT NULL DEFAULT 0,
	confusion_def int4 NOT NULL DEFAULT 0,
	wind int4 NOT NULL DEFAULT 0,
	wind_buff int4 NOT NULL DEFAULT 0,
	water int4 NOT NULL DEFAULT 0,
	water_buff int4 NOT NULL DEFAULT 0,
	fire int4 NOT NULL DEFAULT 0,
	fire_buff int4 NOT NULL DEFAULT 0,
	CONSTRAINT stats_pkey PRIMARY KEY (id)
);

CREATE TABLE hops.users (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	user_name varchar(18) NOT NULL,
	"password" bytea NOT NULL,
	user_type int2 NOT NULL DEFAULT 0,
	ip text NULL DEFAULT ''::text,
	"server" int2 NOT NULL DEFAULT 0,
	ncash int8 NOT NULL DEFAULT 0,
	bank_gold int8 NOT NULL DEFAULT 0,
	mail text NOT NULL DEFAULT ''::text,
	created_at timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+02'::timestamp with time zone,
	disabled_until timestamptz NULL,
	CONSTRAINT users_pkey PRIMARY KEY (id)
);