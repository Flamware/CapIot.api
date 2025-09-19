-- Ce script crée un schéma de base de données complet pour un système de gestion d'appareils.
-- Il gère les appareils, les composants, les emplacements, les utilisateurs et les calendriers de fonctionnement récurrents.
-- Une logique de suppression en cascade garantit la cohérence des données.

-- ====================================================================================================
-- Nettoyage du schéma (en ordre de dépendance)
-- ====================================================================================================

DROP TABLE IF EXISTS public.component_log;
DROP TABLE IF EXISTS public.components;
DROP TABLE IF EXISTS public.device_location;
DROP TABLE IF EXISTS public.user_site;
DROP TABLE IF EXISTS public.locations;
DROP TABLE IF EXISTS public.sites;
DROP TABLE IF EXISTS public.recurring_schedules;
DROP TABLE IF EXISTS public.devices;
DROP TABLE IF EXISTS public.users;


-- ====================================================================================================
-- Schéma Principal
-- ====================================================================================================

-- Table pour les utilisateurs du système.
CREATE TABLE IF NOT EXISTS public.users (
                                            id SERIAL PRIMARY KEY,
                                            name VARCHAR(255),
    email VARCHAR(255) NOT NULL UNIQUE,
    auth0_id VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

-- Table pour les appareils physiques.
CREATE TABLE IF NOT EXISTS public.devices (
                                              device_id TEXT PRIMARY KEY,
                                              last_seen TIMESTAMP WITH TIME ZONE,
                                              status VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                             );

-- Table pour les sites géographiques.
CREATE TABLE IF NOT EXISTS public.sites (
                                            site_id SERIAL PRIMARY KEY,
                                            site_name TEXT NOT NULL,
                                            site_address TEXT,
                                            latitude DECIMAL(9, 6),
    longitude DECIMAL(9, 6)
    );

-- Table pour les emplacements au sein des sites.
CREATE TABLE IF NOT EXISTS public.locations (
                                                location_id SERIAL PRIMARY KEY,
                                                location_name TEXT NOT NULL,
                                                location_description TEXT,
                                                site_id INTEGER,
                                                CONSTRAINT fk_site FOREIGN KEY (site_id) REFERENCES public.sites(site_id) ON DELETE CASCADE
    );

-- Table pour les composants d'un appareil. La suppression d'un appareil supprime ses composants.
CREATE TABLE IF NOT EXISTS public.components (
                                                 component_id TEXT PRIMARY KEY,
                                                 device_id TEXT NOT NULL,
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

-- Lie les utilisateurs à leurs sites assignés.
CREATE TABLE IF NOT EXISTS public.user_site (
                                                user_id INTEGER NOT NULL,
                                                site_id INTEGER NOT NULL,
                                                assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                                PRIMARY KEY (user_id, site_id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE,
    CONSTRAINT fk_site FOREIGN KEY (site_id) REFERENCES public.sites(site_id) ON DELETE CASCADE
    );

-- Suit les emplacements actuels et historiques des appareils.
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

-- Index unique pour s'assurer qu'un appareil n'a qu'un seul emplacement actuel.
CREATE UNIQUE INDEX IF NOT EXISTS unique_current_device_location ON public.device_location (device_id) WHERE is_current = true;

-- ====================================================================================================
-- Tables de Gestion des Calendriers (Règles de Recurrence)
-- ====================================================================================================

-- Table pour stocker les règles de calendrier récurrentes.
-- Utilise le format RRULE iCalendar pour la flexibilité.
CREATE TABLE IF NOT EXISTS public.recurring_schedules (
                                                          recurring_schedule_id SERIAL PRIMARY KEY,
                                                          device_id TEXT NOT NULL,
                                                          schedule_name VARCHAR(255),
    is_exception BOOLEAN DEFAULT FALSE,
    priority INTEGER DEFAULT 0,
    start_time TIME WITH TIME ZONE NOT NULL,
    end_time TIME WITH TIME ZONE NOT NULL,
                      start_date DATE,
                      end_date DATE,
                      recurrence_rule TEXT NOT NULL,
                      CONSTRAINT fk_device_recurring FOREIGN KEY (device_id) REFERENCES public.devices(device_id) ON DELETE CASCADE
    );

-- ====================================================================================================
-- Journalisation et Privilèges
-- ====================================================================================================

-- Table pour stocker les données générées par les composants.
-- La suppression d'un composant supprime ses journaux.
CREATE TABLE IF NOT EXISTS public.component_log (
                                                    log_id SERIAL PRIMARY KEY,
                                                    component_id TEXT NOT NULL,
                                                    log_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                                    log_content TEXT,
                                                    log_read BOOLEAN DEFAULT FALSE,
                                                    CONSTRAINT fk_component FOREIGN KEY (component_id) REFERENCES public.components (component_id) ON DELETE CASCADE
    );

-- Octroi des privilèges au rôle 'admin' sur les tables et séquences du schéma public.
GRANT SELECT, INSERT, UPDA²TE, DELETE ON ALL TABLES IN SCHEMA public TO admin;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO admin;