CREATE TABLE "data".advanced_fusion (
	item1 int8 NOT NULL,
	item2 int8 NOT NULL,
	count2 int2 NOT NULL,
	item3 int8 NOT NULL,
	count3 int2 NOT NULL,
	special_item int8 NOT NULL,
	special_item_count int2 NOT NULL,
	probability int4 NOT NULL,
	"cost" int8 NOT NULL,
	production int8 NOT NULL,
	destroy_on_fail bool NULL DEFAULT false,
	CONSTRAINT advanced_fusion_pkey PRIMARY KEY (item1)
);

CREATE TABLE "data".buff_icons (
	skill_id int4 NOT NULL,
	icon_id int4 NOT NULL,
	CONSTRAINT buff_icons_pkey PRIMARY KEY (skill_id)
);

CREATE TABLE "data".buff_infections (
	id int4 NOT NULL,
	"name" text NOT NULL,
	poison_def int4 NOT NULL,
	paralysis_def int4 NOT NULL,
	confusion_def int4 NOT NULL,
	base_def int4 NOT NULL,
	additional_def int4 NOT NULL,
	arts_def int4 NOT NULL,
	additional_arts_def int4 NOT NULL,
	max_hp int4 NOT NULL,
	hp_recovery_rate int4 NOT NULL,
	str int4 NOT NULL,
	dex int4 NOT NULL,
	"int" int4 NOT NULL,
	base_hp int4 NOT NULL,
	additional_hp int4 NOT NULL,
	base_atk int4 NOT NULL DEFAULT 0,
	additional_atk int4 NOT NULL DEFAULT 0,
	base_arts_atk int4 NOT NULL DEFAULT 0,
	additional_arts_atk int4 NOT NULL DEFAULT 0,
	CONSTRAINT buff_infections_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".drops (
	id int4 NOT NULL,
	items _int8 NULL,
	probabilities _int4 NULL,
	CONSTRAINT drops_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".exp_table (
	"level" int2 NOT NULL,
	"exp" int8 NOT NULL,
	skill_points int4 NOT NULL DEFAULT 0
);

CREATE TABLE "data".gambling (
	id int4 NOT NULL,
	"cost" int8 NULL DEFAULT 0,
	drop_id int4 NOT NULL,
	CONSTRAINT gambling_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".gates (
	id int4 NOT NULL,
	target_map int2 NULL DEFAULT 0,
	point point NOT NULL,
	CONSTRAINT gates_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".hax_codes (
	id int4 NOT NULL,
	code varchar(2) NOT NULL,
	sale_multiplier int4 NULL DEFAULT 0,
	extraction_multiplier int4 NULL DEFAULT 0,
	extracted_item int4 NULL DEFAULT 0,
	CONSTRAINT hax_codes_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".ht_shop (
	id int8 NOT NULL,
	cash int4 NOT NULL,
	is_active bool NOT NULL DEFAULT true,
	ht_id int4 NOT NULL DEFAULT 0,
	is_new bool NOT NULL DEFAULT false,
	is_popular bool NOT NULL DEFAULT false,
	CONSTRAINT ht_shop_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".item_meltings (
	id int4 NOT NULL,
	melted_items _int4 NULL,
	item_counts _int4 NULL,
	profit_multiplier float4 NULL,
	probability int4 NULL DEFAULT 0,
	"cost" int8 NULL DEFAULT 0,
	special_item int4 NULL DEFAULT 0,
	special_probability int4 NULL DEFAULT 0,
	CONSTRAINT item_meltings_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".items (
	id int8 NOT NULL,
	"name" text NOT NULL,
	uif text NULL,
	"type" int2 NOT NULL,
	ht_type int2 NOT NULL,
	timer_type int2 NULL DEFAULT 0,
	timer int4 NULL DEFAULT 0,
	buy_price int8 NULL DEFAULT 0,
	sell_price int8 NULL DEFAULT 0,
	slot int4 NULL DEFAULT 0,
	min_level int4 NULL DEFAULT 0,
	max_level int4 NULL DEFAULT 0,
	base_def1 int4 NULL DEFAULT 0,
	base_def2 int4 NULL DEFAULT 0,
	base_def3 int4 NULL DEFAULT 0,
	base_min_atk int4 NULL DEFAULT 0,
	base_max_atk int4 NULL DEFAULT 0,
	str int4 NULL DEFAULT 0,
	dex int4 NULL DEFAULT 0,
	"int" int4 NULL DEFAULT 0,
	wind int4 NULL DEFAULT 0,
	water int4 NULL DEFAULT 0,
	fire int4 NULL DEFAULT 0,
	max_hp int4 NULL DEFAULT 0,
	max_chi int4 NULL DEFAULT 0,
	min_atk int4 NULL DEFAULT 0,
	max_atk int4 NULL DEFAULT 0,
	atk_rate int4 NULL DEFAULT 0,
	min_arts_atk int4 NULL DEFAULT 0,
	max_arts_atk int4 NULL DEFAULT 0,
	arts_atk_rate int4 NULL DEFAULT 0,
	def int4 NULL DEFAULT 0,
	def_rate int4 NULL DEFAULT 0,
	arts_def int4 NULL DEFAULT 0,
	arts_def_rate int4 NULL DEFAULT 0,
	accuracy int4 NULL DEFAULT 0,
	dodge int4 NULL DEFAULT 0,
	hp_recovery int4 NULL DEFAULT 0,
	chi_recovery int4 NULL DEFAULT 0,
	holy_water_upg1 int4 NULL DEFAULT 0,
	holy_water_upg2 int4 NULL DEFAULT 0,
	holy_water_upg3 int4 NULL DEFAULT 0,
	holy_water_rate1 int4 NULL DEFAULT 0,
	holy_water_rate2 int4 NULL DEFAULT 0,
	holy_water_rate3 int4 NULL DEFAULT 0,
	character_type int4 NULL DEFAULT 0,
	exp_rate float4 NULL DEFAULT 0,
	drop_rate float4 NULL DEFAULT 0,
	tradable bool NULL DEFAULT false,
	min_upgrade_level int2 NOT NULL DEFAULT 0,
	npc_id int4 NOT NULL DEFAULT 0,
	running_speed float4 NOT NULL DEFAULT 0,
	CONSTRAINT items_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".job_passives (
	id int2 NOT NULL,
	max_hp int4 NULL DEFAULT 0,
	max_chi int4 NULL DEFAULT 0,
	atk int4 NULL DEFAULT 0,
	arts_atk int4 NULL DEFAULT 0,
	def int4 NULL DEFAULT 0,
	arts_def int4 NULL DEFAULT 0,
	accuracy int4 NULL DEFAULT 0,
	dodge int4 NULL DEFAULT 0,
	CONSTRAINT job_passives_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".npc_pos_table (
	id int4 NOT NULL,
	npc_id int4 NOT NULL,
	"map" int2 NOT NULL,
	rotation float4 NOT NULL,
	min_location point NOT NULL,
	max_location point NOT NULL,
	count int2 NULL DEFAULT 1,
	respawn_time int4 NOT NULL,
	is_npc bool NOT NULL DEFAULT false,
	attackable bool NOT NULL DEFAULT false,
	CONSTRAINT npc_pos_table_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".npc_scripts (
	id int4 NOT NULL,
	script jsonb NOT NULL,
	CONSTRAINT npc_script_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".npc_table (
	id int4 NOT NULL,
	"name" text NOT NULL,
	"level" int2 NOT NULL,
	"exp" int8 NOT NULL,
	divine_exp int8 NOT NULL,
	darkness_exp int8 NOT NULL,
	gold_drop int4 NOT NULL,
	def int4 NOT NULL,
	max_hp int4 NOT NULL,
	min_atk int4 NOT NULL,
	max_atk int4 NOT NULL,
	arts_def int4 NOT NULL,
	drop_id int4 NOT NULL DEFAULT 0,
	skill_id int4 NOT NULL DEFAULT 0,
	min_arts_atk int4 NOT NULL DEFAULT 0,
	max_arts_atk int4 NOT NULL DEFAULT 0,
	CONSTRAINT npc_table_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".pet_exp_table (
	"level" int2 NOT NULL,
	req_exp_evo1 int4 NOT NULL DEFAULT 0,
	req_exp_evo2 int4 NOT NULL DEFAULT 0,
	req_exp_evo3 int4 NOT NULL DEFAULT 0,
	req_exp_ht int4 NOT NULL DEFAULT 0,
	req_exp_div_evo1 int4 NOT NULL DEFAULT 0,
	req_exp_div_evo2 int4 NOT NULL DEFAULT 0,
	req_exp_div_evo3 int4 NOT NULL DEFAULT 0
);

CREATE TABLE "data".pets (
	id int8 NOT NULL,
	"name" text NOT NULL,
	evolution int2 NOT NULL,
	"level" int2 NOT NULL,
	target_level int2 NOT NULL,
	evolved_id int8 NOT NULL,
	base_str int4 NOT NULL,
	additional_str int4 NOT NULL,
	base_dex int4 NOT NULL,
	additional_dex int4 NOT NULL,
	base_int int4 NOT NULL,
	additional_int int4 NOT NULL,
	base_hp int4 NOT NULL,
	additional_hp int4 NOT NULL,
	base_chi int4 NOT NULL,
	additional_chi int4 NOT NULL,
	skill_id int4 NOT NULL DEFAULT 0,
	combat bool NOT NULL DEFAULT false,
	CONSTRAINT pets_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".productions (
	id int4 NOT NULL,
	materials jsonb NOT NULL,
	probability int4 NOT NULL,
	"cost" int8 NULL DEFAULT 0,
	production int4 NOT NULL,
	CONSTRAINT productions_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".save_points (
	id int2 NOT NULL,
	point point NOT NULL,
	CONSTRAINT save_points_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".shop_items (
	"type" int4 NOT NULL,
	items _int8 NOT NULL,
	CONSTRAINT shop_items_pkey PRIMARY KEY (type)
);

CREATE TABLE "data".shop_table (
	id int4 NOT NULL,
	"name" text NOT NULL,
	"types" _int4 NOT NULL,
	CONSTRAINT shop_table_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".skills (
	id int4 NOT NULL,
	book_id int8 NOT NULL,
	"name" text NOT NULL,
	target int2 NULL DEFAULT 0,
	passive_type int2 NULL DEFAULT 0,
	"type" int2 NULL DEFAULT 0,
	max_plus int2 NULL DEFAULT 0,
	slot int4 NULL DEFAULT 0,
	base_duration int4 NULL DEFAULT 0,
	additional_duration int4 NULL DEFAULT 0,
	cast_time float4 NULL DEFAULT 0,
	base_chi int4 NULL DEFAULT 0,
	additional_chi int4 NULL DEFAULT 0,
	base_min_multiplier int4 NULL DEFAULT 0,
	additional_min_multiplier int4 NULL DEFAULT 0,
	base_max_multiplier int4 NULL DEFAULT 0,
	additional_max_multiplier int4 NULL DEFAULT 0,
	base_radius float4 NULL DEFAULT 0,
	additional_radius float4 NULL DEFAULT 0,
	passive bool NULL DEFAULT false,
	base_passive int4 NULL DEFAULT 0,
	additional_passive int4 NULL DEFAULT 0,
	infection_id int4 NOT NULL DEFAULT 0,
	area_center int4 NOT NULL DEFAULT 0,
	cooldown float4 NOT NULL DEFAULT 0,
	CONSTRAINT skills_pkey PRIMARY KEY (id)
);

CREATE TABLE "data".stackables (
	id int4 NOT NULL,
	uif text NOT NULL,
	CONSTRAINT stackables_pkey PRIMARY KEY (id)
);