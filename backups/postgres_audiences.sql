PGDMP     *    &                |         
   reports_db     15.10 (Debian 15.10-1.pgdg120+1)    15.3     ;           0    0    ENCODING    ENCODING        SET client_encoding = 'UTF8';
                      false            <           0    0 
   STDSTRINGS 
   STDSTRINGS     (   SET standard_conforming_strings = 'on';
                      false            =           0    0 
   SEARCHPATH 
   SEARCHPATH     8   SELECT pg_catalog.set_config('search_path', '', false);
                      false            >           1262    16426 
   reports_db    DATABASE     u   CREATE DATABASE reports_db WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.utf8';
    DROP DATABASE reports_db;
                postgres    false            S           1247    16518    cabinet_name_enum    TYPE     ]   CREATE TYPE public.cabinet_name_enum AS ENUM (
    'yandex',
    'google',
    'facebook'
);
 $   DROP TYPE public.cabinet_name_enum;
       public          postgres    false            �            1255    16495    update_updated_at_column()    FUNCTION     �   CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;
 1   DROP FUNCTION public.update_updated_at_column();
       public          postgres    false            �            1259    16573    audience_filters_id_seq    SEQUENCE     �   CREATE SEQUENCE public.audience_filters_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 .   DROP SEQUENCE public.audience_filters_id_seq;
       public          postgres    false            �            1259    16564    audience_filters    TABLE     �  CREATE TABLE public.audience_filters (
    id integer DEFAULT nextval('public.audience_filters_id_seq'::regclass) NOT NULL,
    audience_id bigint NOT NULL,
    creation_date_from timestamp without time zone NOT NULL,
    creation_date_to timestamp without time zone NOT NULL,
    status_ids integer[],
    reason_ids integer[],
    status_names text[],
    non_target_reasons text[],
    rejection_reasons text[]
);
 $   DROP TABLE public.audience_filters;
       public         heap    postgres    false    221            �            1259    16437    audience_requests    TABLE     W  CREATE TABLE public.audience_requests (
    id integer NOT NULL,
    audience_id integer,
    request_id integer NOT NULL,
    status character varying(50),
    reason character varying(255),
    creation_date_from timestamp without time zone,
    creation_date_to timestamp without time zone,
    client_id integer,
    manager_id integer
);
 %   DROP TABLE public.audience_requests;
       public         heap    postgres    false            �            1259    16436    audience_requests_id_seq    SEQUENCE     �   CREATE SEQUENCE public.audience_requests_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 /   DROP SEQUENCE public.audience_requests_id_seq;
       public          postgres    false    217            ?           0    0    audience_requests_id_seq    SEQUENCE OWNED BY     U   ALTER SEQUENCE public.audience_requests_id_seq OWNED BY public.audience_requests.id;
          public          postgres    false    216            �            1259    16428 	   audiences    TABLE       CREATE TABLE public.audiences (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);
    DROP TABLE public.audiences;
       public         heap    postgres    false            �            1259    16427    audiences_id_seq    SEQUENCE     �   CREATE SEQUENCE public.audiences_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 '   DROP SEQUENCE public.audiences_id_seq;
       public          postgres    false    215            @           0    0    audiences_id_seq    SEQUENCE OWNED BY     E   ALTER SEQUENCE public.audiences_id_seq OWNED BY public.audiences.id;
          public          postgres    false    214            �            1259    16548    integrations_integration_id_seq    SEQUENCE     �   CREATE SEQUENCE public.integrations_integration_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 6   DROP SEQUENCE public.integrations_integration_id_seq;
       public          postgres    false            �            1259    16549    integrations    TABLE     �  CREATE TABLE public.integrations (
    id bigint DEFAULT nextval('public.integrations_integration_id_seq'::regclass) NOT NULL,
    audience_id integer NOT NULL,
    cabinet_name public.cabinet_name_enum NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    external_id integer
);
     DROP TABLE public.integrations;
       public         heap    postgres    false    218    851            �           2604    16440    audience_requests id    DEFAULT     |   ALTER TABLE ONLY public.audience_requests ALTER COLUMN id SET DEFAULT nextval('public.audience_requests_id_seq'::regclass);
 C   ALTER TABLE public.audience_requests ALTER COLUMN id DROP DEFAULT;
       public          postgres    false    216    217    217            �           2604    16431    audiences id    DEFAULT     l   ALTER TABLE ONLY public.audiences ALTER COLUMN id SET DEFAULT nextval('public.audiences_id_seq'::regclass);
 ;   ALTER TABLE public.audiences ALTER COLUMN id DROP DEFAULT;
       public          postgres    false    215    214    215            �           2606    16570 &   audience_filters audience_filters_pkey 
   CONSTRAINT     d   ALTER TABLE ONLY public.audience_filters
    ADD CONSTRAINT audience_filters_pkey PRIMARY KEY (id);
 P   ALTER TABLE ONLY public.audience_filters DROP CONSTRAINT audience_filters_pkey;
       public            postgres    false    220            �           2606    16572    audience_filters audience_id 
   CONSTRAINT     ^   ALTER TABLE ONLY public.audience_filters
    ADD CONSTRAINT audience_id UNIQUE (audience_id);
 F   ALTER TABLE ONLY public.audience_filters DROP CONSTRAINT audience_id;
       public            postgres    false    220            �           2606    16442 (   audience_requests audience_requests_pkey 
   CONSTRAINT     f   ALTER TABLE ONLY public.audience_requests
    ADD CONSTRAINT audience_requests_pkey PRIMARY KEY (id);
 R   ALTER TABLE ONLY public.audience_requests DROP CONSTRAINT audience_requests_pkey;
       public            postgres    false    217            �           2606    16435    audiences audiences_pkey 
   CONSTRAINT     V   ALTER TABLE ONLY public.audiences
    ADD CONSTRAINT audiences_pkey PRIMARY KEY (id);
 B   ALTER TABLE ONLY public.audiences DROP CONSTRAINT audiences_pkey;
       public            postgres    false    215            �           2606    16556    integrations integrations_pkey 
   CONSTRAINT     \   ALTER TABLE ONLY public.integrations
    ADD CONSTRAINT integrations_pkey PRIMARY KEY (id);
 H   ALTER TABLE ONLY public.integrations DROP CONSTRAINT integrations_pkey;
       public            postgres    false    219            �           2606    16516    audiences uniq_name 
   CONSTRAINT     N   ALTER TABLE ONLY public.audiences
    ADD CONSTRAINT uniq_name UNIQUE (name);
 =   ALTER TABLE ONLY public.audiences DROP CONSTRAINT uniq_name;
       public            postgres    false    215            �           1259    16463 !   idx_audience_requests_audience_id    INDEX     f   CREATE INDEX idx_audience_requests_audience_id ON public.audience_requests USING btree (audience_id);
 5   DROP INDEX public.idx_audience_requests_audience_id;
       public            postgres    false    217            �           1259    16562    idx_integrations_audience_id    INDEX     \   CREATE INDEX idx_integrations_audience_id ON public.integrations USING btree (audience_id);
 0   DROP INDEX public.idx_integrations_audience_id;
       public            postgres    false    219            �           2620    16563 +   integrations update_integrations_updated_at    TRIGGER     �   CREATE TRIGGER update_integrations_updated_at BEFORE UPDATE ON public.integrations FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
 D   DROP TRIGGER update_integrations_updated_at ON public.integrations;
       public          postgres    false    222    219            �           2606    16443 4   audience_requests audience_requests_audience_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.audience_requests
    ADD CONSTRAINT audience_requests_audience_id_fkey FOREIGN KEY (audience_id) REFERENCES public.audiences(id);
 ^   ALTER TABLE ONLY public.audience_requests DROP CONSTRAINT audience_requests_audience_id_fkey;
       public          postgres    false    3227    217    215            �           2606    16557    integrations fk_audience    FK CONSTRAINT     �   ALTER TABLE ONLY public.integrations
    ADD CONSTRAINT fk_audience FOREIGN KEY (audience_id) REFERENCES public.audiences(id) ON DELETE CASCADE;
 B   ALTER TABLE ONLY public.integrations DROP CONSTRAINT fk_audience;
       public          postgres    false    219    215    3227           