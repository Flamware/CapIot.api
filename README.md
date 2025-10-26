## 🚀 README: API Project

Ce README fournit les étapes nécessaires pour configurer et exécuter le projet API.

-----

### **⚠️ Note importante pour les utilisateurs Windows**

Pour garantir une **compatibilité maximale** avec l'environnement de développement et les outils comme Flyway, il est fortement recommandé d'utiliser **WSL2 (Windows Subsystem for Linux 2)** pour exécuter toutes les commandes (Git, Go, Flyway) et pour travailler sur les fichiers du projet.

-----

### **Prerequisites**

* **Go** (Golang) installé sur votre système (de préférence dans l'environnement WSL2).
* **PostgreSQL** server running locally or accessible.
* **Flyway** command-line tool installed (used for database migrations).
* **MQTT Broker** (comme Mosquitto) running on port `1883`.
* **Git** installed.

-----

### **Setup Instructions**

#### **1. Switch to the Development Branch**

Avant de commencer, assurez-vous d'être sur la bonne branche de développement :

```bash
git checkout DEV
```

-----

#### **2. IDE Setup (IntelliJ Users)**

Si vous utilisez IntelliJ IDEA, assurez-vous que le **Go Language Plugin** est installé pour supporter le développement Go.

-----

#### **3. Environment Configuration**

Vous devez avoir un fichier **`.env`** à la racine du dossier `api/cmd`. Ce fichier contiendra les variables de configuration nécessaires, notamment pour la base de données et le broker MQTT :

> **IMPORTANT :** Le fichier `.env` doit contenir la variable pour le broker MQTT, qui doit être accessible via `localhost:1883` depuis votre environnement WSL2.
>
> *Exemple de variable `.env` :*
>
> ```env
> MQTT_BROKER=localhost:1883
> # ... autres variables de DB, etc.
> ```

-----

#### **4. PostgreSQL Setup on WSL2**

Installez PostgreSQL directement dans votre distribution Linux (exemple Ubuntu/Debian) :

1.  **Installer et démarrer PostgreSQL :**
    ```bash
    sudo apt update
    sudo apt install postgresql postgresql-contrib
    sudo service postgresql start
    ```
2.  **Configuration de l'utilisateur de l'application (`admin`) :**
    * Créez/modifiez l'utilisateur (`admin`) et définissez un mot de passe :
      ```bash
      sudo -u postgres psql
      CREATE USER admin WITH ENCRYPTED PASSWORD 'VOTRE_MOT_DE_PASSE';
      \q
      ```

-----

#### **5. MQTT Broker Setup**

Le service nécessite une connexion à un broker MQTT sur le port `1883`.

1.  **Installation de Mosquitto (Broker MQTT) sur WSL2 :**

    ```bash
    sudo apt install mosquitto mosquitto-clients
    ```

2.  **Démarrage du service Mosquitto :**

    ```bash
    sudo service mosquitto start
    ```

    > **Dépannage de l'erreur `connection refused` (127.0.0.1:1883) :**

    > Cette erreur indique que le broker Mosquitto n'était pas démarré ou n'écoutait pas sur l'interface `127.0.0.1`. Assurez-vous que l'étape 2 (`sudo service mosquitto start`) a réussi avant de lancer l'API.

-----

#### **6. Database Setup and Migration**

* **Create the Database:** Créez la base de données **`Capiot`** :

  ```bash
  sudo -u postgres createdb capiot
  ```

* **Grant Privileges:** Accordez à l'utilisateur de l'application (`admin`) les droits nécessaires :

    1.  **Accès à la base de données :**
        ```bash
        sudo -u postgres psql
        GRANT ALL PRIVILEGES ON DATABASE capiot TO admin;
        \q
        ```
    2.  **Droits de création sur le schéma `public` (pour Flyway) :**
        ```bash
        sudo -u postgres psql -d capiot
        GRANT CREATE ON SCHEMA public TO admin;
        \q
        ```

* **Database Migration (Flyway):** Utilisez **Flyway** pour appliquer les migrations.

  Exécutez la commande Flyway :

  ```bash
  flyway migrate
  ```

-----

### **Running the API**

1.  **Download Dependencies:**
    ```bash
    go mod download
    ```
2.  **Run the Application:**
    ```bash
    go run server.go
    ```

L'API devrait maintenant se connecter à la base de données et au broker MQTT.

-----