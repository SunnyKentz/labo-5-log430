const fetch = (...args) => import('node-fetch').then(({ default: fetch }) => fetch(...args));
let magasin = 'Magasin 1';
let caisse = 'Caisse 1';
loadTestRapport()
async function loadTestRapport() {
    //  Génération de rapports consolidés.
    let token = await loginAndGetToken('http://localhost/mere/api/v1/merelogin', magasin, caisse)
    fetch('http://localhost/mere/api/v1/raport', {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${token}` }
    })
}

async function loginAndGetToken(loginEndpoint, magasin, caisse) {
    const loginPayload = {
        username: 'Bob',
        password: 'password',
        magasin: magasin,
        caisse: caisse
    };
    try {
        const res = await fetch(loginEndpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(loginPayload)
        });
        if (!res.ok) {
            throw new Error(`Login failed: ${res.status} ${res.statusText}`);
        }
        const data = await res.json();
        return data.token || data.access_token || null;
    } catch (err) {
        console.error('Error during login:', err.message,loginEndpoint);
        return null;
    }
}