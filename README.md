## Labo 5

### Explication de l'application:

l'application sur loacalhost:80 avec des routes différent pour chaques services.

On peut login avec Bob en tant que manager et Alice en tant que commis
le nom du magasin peut etre nimporte quoi
la caisse doivent etre Caisse 1, Caisse 2 ou Caisse 3

le mdp est toujour : password

Nous pouvons enregistrer un nouveau user/client dans loacalhost/mere

### graphana :
lien : `localhost:3000`


### Comment run :
```
    make run
```

### Comment tester :
```
    make test
```

### Comment generer les docs :
```
    make docs
```

### Explication du CI
Après avoir fait un push, Github action check le linting du push,<br> execute les testes et si tous passe, le push vers dockerhub
