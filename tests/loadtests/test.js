const axios = require('axios');
let magasin = 'Magasin 1';
let caisse = 'Caisse 1';

main()

function main() {
    const argv = process.argv;
    const rps = parseInt(argv[2], 10) || 10; // Default to 10 rps if not provided
    console.log(`Starting load test with ${rps} requests per second...`);
    loadTestConsultation(rps)
    loadTestMiseAJour(rps)
    loadTestRapport(rps)
}

async function loadTestConsultation(requestsPerSecond) {
    // Consultation simultanée des stocks de plusieurs magasins.
    let initial = requestsPerSecond
    let token = await loginAndGetToken('http://localhost:57463/magasin/api/v1/login', magasin, caisse)
    setInterval(() => {
        let count = requestsPerSecond;
        while (count > 0) {
            axios.get('http://localhost:57463/magasin/api/v1/produits', {
                headers: { 'Authorization': `Bearer ${token}`, 'C-Mag': magasin, 'C-Caisse': caisse }
            }).catch(() => {});
            count--;
        }
    }, 1000);
}

async function loadTestRapport(requestsPerSecond) {
    //  Génération de rapports consolidés.
    let token = await loginAndGetToken('http://localhost:57463/mere/api/v1/merelogin', magasin, caisse)
    let initial = requestsPerSecond
    setInterval(() => {
        let count = requestsPerSecond;
        while (count > 0) {
            axios.get('http://localhost:57463/mere/api/v1/raport', {
                headers: { 'Authorization': `Bearer ${token}` }
            }).catch(() => {});
            count--;
        }
    }, 1000);
}

async function loadTestMiseAJour(requestsPerSecond) {
    //  Mise à jour de produits à forte fréquence. 
    let token = await loginAndGetToken('http://localhost:57463/mere/api/v1/merelogin', magasin, caisse)
    let initial = requestsPerSecond
    const productData = {
        productId: 5,
        nom: "new nom",
        prix: 23.45,
        description: "description",
    };
    setInterval(() => {
        let count = requestsPerSecond;
        while (count > 0) {
            axios.put('http://localhost:57463/mere/api/v1/produit', productData, {
                headers: { 'Authorization': `Bearer ${token}`,"Content-Type": "application/json" }
            }).catch(() => {});
            count--;
        }
    }, 1000);
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
        console.error('Error during login:', err.message,loginEndpoint);
        return null;
    }
}