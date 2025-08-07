const axios = require('axios');
let magasin = 'Magasin 1';
let caisse = 'Caisse 1';

loadTestMiseAJour();

async function loadTestMiseAJour() {
    //  Mise à jour de produits à forte fréquence. 
    let token = await loginAndGetToken('http://localhost:61867/mere/api/v1/merelogin', magasin, caisse);

    const productData = {
        productId: 5,
        nom: "new nom",
        prix: 23.45,
        description: "description",
    };

    try {
        await axios.put(
            'http://localhost:61867/mere/api/v1/produit',
            productData,
            {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            }
        );
    } catch (err) {
        console.error('Error during product update:', err.message);
    }
}

async function loginAndGetToken(loginEndpoint, magasin, caisse) {
    const loginPayload = {
        username: 'Bob',
        password: 'password',
        magasin: magasin,
        caisse: caisse
    };
    try {
        const res = await axios.post(loginEndpoint, loginPayload, {
            headers: { 'Content-Type': 'application/json' }
        });
        const data = res.data;
        return data.token || data.access_token || null;
    } catch (err) {
        console.error('Error during login:', err.message, loginEndpoint);
        return null;
    }
}