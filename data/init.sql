
-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    magasin VARCHAR(100) NOT NULL,
    caisse VARCHAR(100) NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('VENTE', 'RETOUR')),
    produit_ids VARCHAR(100) NOT NULL,
    montant DECIMAL(10,2) NOT NULL,
    deja_retourne BOOLEAN NOT NULL DEFAULT FALSE
);

-- Create caisse table
CREATE TABLE IF NOT EXISTS caisses (
    id SERIAL PRIMARY KEY,
    nom VARCHAR(100) NOT NULL,
    occupe BOOLEAN NOT NULL DEFAULT FALSE
);
-- Insert caisses
INSERT INTO caisses (nom, occupe) VALUES
('Caisse 1', FALSE),
('Caisse 2', FALSE),
('Caisse 3', FALSE);

-- Create employe table
CREATE TABLE IF NOT EXISTS employes (
    id SERIAL PRIMARY KEY,
    nom VARCHAR(100) NOT NULL,
    role VARCHAR(100) NOT NULL
);
-- Insert sample employees
INSERT INTO employes (nom, role) VALUES
('Alice', 'commis'),
('Bob', 'manager'),
('Claire', 'commis'),
('David', 'commis'),
('Eva', 'manager');


CREATE TABLE IF NOT EXISTS demande_reaprovisionements (
    id SERIAL PRIMARY KEY,
    produit_id INTEGER NOT NULL,
    produit_nom VARCHAR(100) NOT NULL,
    quantite INTEGER NOT NULL
);

-- Create products table
CREATE TABLE IF NOT EXISTS produits (
    id SERIAL PRIMARY KEY,
    nom VARCHAR(100) NOT NULL,
    description TEXT,
    prix DECIMAL(10,2) NOT NULL,
    categorie VARCHAR(50) NOT NULL,
    quantite INTEGER NOT NULL DEFAULT 0
);

-- Insert sample products
INSERT INTO produits (nom, prix, description, categorie, quantite) VALUES

-- Boissons
('Eau minérale 1.5L', 1.20, 'Bouteille d''eau minérale naturelle 1,5 litre', 'Boissons', 8),
('Coca-Cola 2L', 2.50, 'Soda Coca-Cola bouteille 2 litres', 'Boissons', 5),
('Jus d''orange 1L', 2.80, 'Jus d''orange pur jus 1 litre', 'Boissons', 7),
('Café en grains 250g', 4.50, 'Café en grains torréfié 250g', 'Boissons', 4),
('Thé vert 100 sachets', 3.90, 'Boîte de 100 sachets de thé vert', 'Boissons', 6),
('Bière blonde 6x33cl', 5.00, 'Pack de 6 bières blondes 33cl', 'Boissons', 3),
('Vin rouge 75cl', 8.50, 'Bouteille de vin rouge 75cl', 'Boissons', 5),
('Eau gazeuse 1L', 1.00, 'Bouteille d''eau gazeuse 1 litre', 'Boissons', 9),
('Limonade 1.5L', 1.70, 'Bouteille de limonade 1,5 litre', 'Boissons', 7),
('Thé glacé 500ml', 1.50, 'Bouteille de thé glacé 500ml', 'Boissons', 8),

-- Produits laitiers
('Lait entier 1L', 1.30, 'Brique de lait entier 1 litre', 'Produits laitiers', 6),
('Yaourt nature 4x125g', 2.20, 'Pack de 4 yaourts nature 125g', 'Produits laitiers', 8),
('Fromage râpé 200g', 2.80, 'Sachet de fromage râpé 200g', 'Produits laitiers', 5),
('Beurre doux 250g', 2.50, 'Plaquette de beurre doux 250g', 'Produits laitiers', 7),
('Crème fraîche 20cl', 1.80, 'Pot de crème fraîche 20cl', 'Produits laitiers', 4),
('Lait demi-écrémé 1L', 1.10, 'Brique de lait demi-écrémé 1 litre', 'Produits laitiers', 9),
('Fromage à tartiner 200g', 2.60, 'Fromage à tartiner nature 200g', 'Produits laitiers', 6),
('Crème dessert chocolat 4x100g', 3.00, 'Pack de 4 crèmes dessert chocolat 100g', 'Produits laitiers', 5),
('Lait en poudre 400g', 4.20, 'Boîte de lait en poudre 400g', 'Produits laitiers', 3),
('Yaourt aux fruits 4x125g', 2.50, 'Pack de 4 yaourts aux fruits 125g', 'Produits laitiers', 7),

-- Épicerie
('Pâtes spaghetti 500g', 1.20, 'Paquet de spaghetti 500g', 'Épicerie', 8),
('Riz basmati 1kg', 2.30, 'Sachet de riz basmati 1kg', 'Épicerie', 6),
('Huile d''olive 1L', 5.90, 'Bouteille d''huile d''olive 1 litre', 'Épicerie', 4),
('Sucre en poudre 1kg', 1.50, 'Sachet de sucre en poudre 1kg', 'Épicerie', 9),
('Sel fin 1kg', 0.90, 'Sachet de sel fin 1kg', 'Épicerie', 10),
('Farine de blé 1kg', 1.00, 'Sachet de farine de blé 1kg', 'Épicerie', 7),
('Tomates pelées 400g', 1.20, 'Boîte de tomates pelées 400g', 'Épicerie', 5),
('Sauce tomate 500g', 2.00, 'Pot de sauce tomate 500g', 'Épicerie', 6),
('Céréales petit-déjeuner 500g', 3.50, 'Boîte de céréales petit-déjeuner 500g', 'Épicerie', 4),
('Lentilles vertes 500g', 1.80, 'Sachet de lentilles vertes 500g', 'Épicerie', 5),

-- Snacks
('Chips nature 150g', 1.80, 'Sachet de chips nature 150g', 'Snacks', 9),
('Biscuits chocolat 300g', 2.40, 'Paquet de biscuits au chocolat 300g', 'Snacks', 7),
('Barres chocolatées x5', 3.20, 'Pack de 5 barres chocolatées', 'Snacks', 6),
('Pop-corn 200g', 2.10, 'Sachet de pop-corn 200g', 'Snacks', 5),
('Crackers 200g', 1.90, 'Sachet de crackers 200g', 'Snacks', 8),
('Noix de cajou 150g', 3.00, 'Sachet de noix de cajou grillées 150g', 'Snacks', 4),
('Amandes grillées 200g', 4.00, 'Sachet d''amandes grillées 200g', 'Snacks', 3),
('Popcorn sucré 100g', 1.50, 'Sachet de popcorn sucré 100g', 'Snacks', 10),
('Gâteaux apéritifs 150g', 2.20, 'Sachet de gâteaux apéritifs 150g', 'Snacks', 8),
('Barres de céréales x6', 2.80, 'Pack de 6 barres de céréales', 'Snacks', 6),

-- Hygiène
('Dentifrice 100ml', 2.30, 'Tube de dentifrice 100ml', 'Hygiène', 5),
('Savon liquide 300ml', 3.50, 'Flacon de savon liquide 300ml', 'Hygiène', 4),
('Shampoing 400ml', 3.20, 'Bouteille de shampoing 400ml', 'Hygiène', 6),
('Déodorant 150ml', 3.80, 'Spray déodorant 150ml', 'Hygiène', 7),
('Papier toilette 8 rouleaux', 2.90, 'Pack de 8 rouleaux de papier toilette', 'Hygiène', 8),
('Gel douche 250ml', 3.00, 'Flacon de gel douche 250ml', 'Hygiène', 5),
('Crème hydratante 200ml', 4.50, 'Pot de crème hydratante 200ml', 'Hygiène', 4),
('Brosse à dents souple', 1.50, 'Brosse à dents à poils souples', 'Hygiène', 9),
('Coton-tiges x200', 2.00, 'Boîte de 200 coton-tiges', 'Hygiène', 11),
('Lingettes nettoyantes x20', 2.50, 'Paquet de 20 lingettes nettoyantes', 'Hygiène', 6);

