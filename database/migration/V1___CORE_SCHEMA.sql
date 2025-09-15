-- Ce script crée un schéma de base de données complet pour un système de gestion d'appareils.
-- Il intègre une logique de suppression en cascade pour garantir que lorsqu'un appareil
-- est supprimé, tous les composants et les journaux de ces composants le sont aussi.

-- ====================================================================================================
-- Nettoyage du schéma (en ordre de dépendance)
-- ====================================================================================================

DROP TABLE IF EXISTS public.component_log;
DROP TABLE IF EXISTS public.components; -- Supprimé en premier car il sera recréé avec la clé étrangère vers `devices`
DROP TABLE IF EXISTS public.device_location;
DROP TABLE IF EXISTS public.user_site;
DROP TABLE IF EXISTS public.locations;
DROP TABLE IF EXISTS public.sites;
DROP TABLE IF EXISTS public.devices;
DROP TABLE IF EXISTS public.users;


-- ====================================================================================================
-- Schéma Principal
-- ====================================================================================================

-- Création de la table des utilisateurs
CREATE TABLE IF NOT EXISTS public.users (
                                            id SERIAL PRIMARY KEY,
                                            name VARCHAR(255),
    email VARCHAR(255) NOT NULL UNIQUE,
    auth0_id VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

-- Création de la table des appareils
CREATE TABLE IF NOT EXISTS public.devices (
                                              device_id TEXT PRIMARY KEY,
                                              last_seen TIMESTAMP WITH TIME ZONE,
                                              status VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                             );

-- Création de la table des sites
CREATE TABLE IF NOT EXISTS public.sites (
                                            site_id SERIAL PRIMARY KEY,
                                            site_name TEXT NOT NULL,
                                            site_address TEXT,
                                            latitude DECIMAL(9, 6),
    longitude DECIMAL(9, 6)
    );

-- Création de la table des emplacements
CREATE TABLE IF NOT EXISTS public.locations (
                                                location_id SERIAL PRIMARY KEY,
                                                location_name TEXT NOT NULL,
                                                location_description TEXT,
                                                site_id INTEGER,
                                                CONSTRAINT fk_site FOREIGN KEY (site_id) REFERENCES public.sites(site_id) ON DELETE CASCADE
    );

-- Création de la table des composants
-- Un composant est maintenant directement lié à un appareil.
-- Si un appareil est supprimé, ses composants seront supprimés en cascade.
CREATE TABLE IF NOT EXISTS public.components (
                                                 component_id TEXT PRIMARY KEY,
                                                 device_id TEXT NOT NULL, -- Nouvelle colonne pour lier directement le composant à un appareil
                                                 component_name VARCHAR(255) NOT NULL,
    component_type VARCHAR(255) NOT NULL,
    component_subtype VARCHAR(255),
    component_status VARCHAR(50),
    min_threshold NUMERIC(10, 2),
    max_threshold NUMERIC(10, 2),
    max_running_hours INTEGER,
    current_running_hours INTEGER DEFAULT 0,
    CONSTRAINT fk_device FOREIGN KEY (device_id) REFERENCES public.devices(device_id) ON DELETE CASCADE
    );

-- ====================================================================================================
-- Tables de Liaison
-- ====================================================================================================

-- Création de la table user_site pour lier les utilisateurs à leurs sites
CREATE TABLE IF NOT EXISTS public.user_site (
                                                user_id INTEGER NOT NULL,
                                                site_id INTEGER NOT NULL,
                                                assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                                PRIMARY KEY (user_id, site_id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE,
    CONSTRAINT fk_site FOREIGN KEY (site_id) REFERENCES public.sites(site_id) ON DELETE CASCADE
    );

-- Création de la table device_location pour suivre les emplacements actuels et historiques d'un appareil
CREATE TABLE IF NOT EXISTS public.device_location (
                                                      id SERIAL PRIMARY KEY,
                                                      device_id TEXT,
                                                      location_id INTEGER,
                                                      assigned_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    is_current BOOLEAN DEFAULT true,
    CONSTRAINT unique_device_location UNIQUE (device_id, location_id),
    CONSTRAINT device_location_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(device_id) ON DELETE CASCADE,
    CONSTRAINT device_location_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(location_id) ON DELETE CASCADE
    );

-- Création d'un index unique pour garantir qu'un appareil ne peut se trouver que dans un seul emplacement actuel
CREATE UNIQUE INDEX IF NOT EXISTS unique_current_device_location ON public.device_location (device_id) WHERE is_current = true;

-- ====================================================================================================
-- Journalisation et Privilèges
-- ====================================================================================================

-- Création de la table component_log pour stocker les données générées par les composants
-- La suppression d'un composant entraînera la suppression de ses journaux.
CREATE TABLE IF NOT EXISTS public.component_log (
                                                    log_id SERIAL PRIMARY KEY,
                                                    component_id TEXT NOT NULL,
                                                    log_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                                    log_content TEXT,
                                                    log_read BOOLEAN DEFAULT FALSE,
                                                    CONSTRAINT fk_component
                                                    FOREIGN KEY (component_id)
    REFERENCES public.components (component_id)
    ON DELETE CASCADE -- Changé de ON DELETE SET NULL à ON DELETE CASCADE
    );

-- Octroi des privilèges au rôle 'admin' sur toutes les tables du schéma public
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO admin;

-- Octroi du privilège USAGE au rôle 'admin' sur toutes les séquences du schéma public
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO admin;
