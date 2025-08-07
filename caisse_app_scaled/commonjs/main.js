function authHeaders() {
    return {
        "Authorization": `Bearer ${localStorage.getItem('jwt')}`,
        "C-Mag": localStorage.getItem('mag'),
        "C-Caisse": localStorage.getItem('caisse')
    }
}