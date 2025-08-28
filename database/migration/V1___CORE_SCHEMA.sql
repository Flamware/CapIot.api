-- Ce script crée un schéma de base de données complet pour un système de gestion d'appareils.
-- Il inclut des tables pour les utilisateurs, les appareils, les emplacements et les composants (capteurs, actionneurs, etc.).
-- Il intègre des fonctionnalités telles que la durée de vie des composants, les dates d'installation et la journalisation des données.

-- ====================================================================================================
-- Schéma Principal
-- ====================================================================================================

-- Création de la table des utilisateurs
CREATE TABLE IF NOT EXISTS public.users (
                                            id SERIAL PRIMARY KEY,
                                            email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    password VARCHAR(255),
    role VARCHAR(50),
    auth0_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

-- Création de la table des appareils
CREATE TABLE IF NOT EXISTS public.devices (
                                              device_id TEXT PRIMARY KEY,
                                              last_seen TIMESTAMP WITH TIME ZONE,
                                              status VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                             );

-- Création de la table des sites (nouveau)
-- Un site est un emplacement physique plus large (ex: un bâtiment, une usine)
-- Chaque site est géré par un client
CREATE TABLE IF NOT EXISTS public.sites (
                                            site_id SERIAL PRIMARY KEY,
                                            site_name TEXT NOT NULL,
                                            site_address TEXT,
                                            latitude DECIMAL(9, 6),
    longitude DECIMAL(9, 6)
    );

-- Création de la table des emplacements (modifié)
-- Une localisation est une pièce ou une zone spécifique à l'intérieur d'un site
CREATE TABLE IF NOT EXISTS public.locations (
                                                location_id SERIAL PRIMARY KEY,
                                                location_name TEXT NOT NULL,
                                                location_description TEXT,
                                                site_id INTEGER, -- Clé étrangère vers la table des sites
                                                CONSTRAINT fk_site FOREIGN KEY (site_id) REFERENCES public.sites(site_id) ON DELETE CASCADE
    );

-- Création de la table des composants physiques (instances)
-- Chaque ligne représente une instance unique d'un composant physique.
-- 'component_id' est son identifiant unique (généré par l'appareil comme une chaîne de texte).
-- 'component_name' est le nom générique du modèle de composant (ex: 'temp-sim-001').
CREATE TABLE IF NOT EXISTS public.components (
                                                 component_id TEXT PRIMARY KEY, -- Identifiant unique pour CHAQUE composant physique
                                                 component_name VARCHAR(255) NOT NULL, -- Nom générique du modèle de composant
    component_type VARCHAR(255) NOT NULL,
    component_subtype VARCHAR(255),
    component_status VARCHAR(50),
    min_threshold NUMERIC(10, 2),
    max_threshold NUMERIC(10, 2),
    max_running_hours INTEGER
    );

-- ====================================================================================================
-- Tables de Liaison
-- ====================================================================================================

-- Création de la table user_location pour lier les utilisateurs à leurs emplacements
CREATE TABLE IF NOT EXISTS public.user_location (
                                                    user_id INTEGER NOT NULL,
                                                    location_id INTEGER NOT NULL,
                                                    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                                    PRIMARY KEY (user_id, location_id),
    CONSTRAINT user_location_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(location_id) ON DELETE CASCADE,
    CONSTRAINT user_location_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
    );

-- Création de la table device_location pour suivre les emplacements actuels et historiques d'un appareil
CREATE TABLE IF NOT EXISTS public.device_location (
                                                      id SERIAL PRIMARY KEY,
                                                      device_id TEXT,
                                                      location_id INTEGER,
                                                      assigned_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    is_current BOOLEAN DEFAULT true,
    CONSTRAINT device_location_device_id_location_id_key UNIQUE (device_id, location_id),
    CONSTRAINT unique_device_location UNIQUE (device_id, location_id),
    CONSTRAINT device_location_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(device_id) ON DELETE CASCADE,
    CONSTRAINT device_location_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(location_id) ON DELETE CASCADE
    );

-- Création d'un index unique pour garantir qu'un appareil ne peut se trouver que dans un seul emplacement actuel
CREATE UNIQUE INDEX IF NOT EXISTS unique_current_device_location ON public.device_location (device_id) WHERE is_current = true;

-- Création de la table device_components pour lier les appareils à leurs composants spécifiques
-- Cette table enregistre l'historique des installations de composants sur les appareils.
-- Un composant ne peut être "actuellement" installé que sur un seul appareil à la fois.
CREATE TABLE IF NOT EXISTS public.device_components (
                                                        id SERIAL PRIMARY KEY, -- Clé primaire auto-incrémentée pour chaque enregistrement d'installation
                                                        device_id TEXT NOT NULL,
                                                        component_id TEXT NOT NULL, -- Identifiant unique du composant physique
                                                        installation_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                                        removal_date TIMESTAMP WITH TIME ZONE, -- Date de retrait du composant
                                                        FOREIGN KEY (device_id) REFERENCES public.devices(device_id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES public.components(component_id) ON DELETE CASCADE,
    UNIQUE (device_id, component_id) -- Contrainte d'unicité pour la paire (appareil, composant)
    );

-- Ajout d'un index unique partiel pour s'assurer qu'un composant n'est "actuellement" installé
-- (c'est-à-dire, removal_date IS NULL) que sur un seul appareil à la fois.
CREATE UNIQUE INDEX IF NOT EXISTS unique_active_component_installation
    ON public.device_components (component_id)
    WHERE removal_date IS NULL;

-- ====================================================================================================
-- Journalisation et Privilèges
-- ====================================================================================================

-- Création de la table component_log pour stocker les données générées par les composants
CREATE TABLE IF NOT EXISTS public.component_log (
                                                    log_id SERIAL PRIMARY KEY,
                                                    component_id TEXT NOT NULL, -- Lien vers le composant physique spécifique
                                                    log_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                                    log_content TEXT,
                                                    log_read BOOLEAN DEFAULT FALSE,
                                                    CONSTRAINT fk_component
                                                    FOREIGN KEY (component_id)
    REFERENCES public.components (component_id)
    ON DELETE SET NULL -- Si un composant est supprimé, les journaux le concernant peuvent rester, mais component_id devient NULL
    );

-- Octroi des privilèges au rôle 'admin' sur toutes les tables du schéma public
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO admin;

-- Octroi du privilège USAGE au rôle 'admin' sur toutes les séquences du schéma public
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO admin;
