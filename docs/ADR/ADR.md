### Justification des décisions d’architecture (ADR)

#### ADR : Choix du protocole HTTP/REST pour la communication inter-services

- **Titre** : Choix de HTTP/REST (JSON) comme protocole de communication entre microservices au lieu de Protobuf/gRPC.
- **Status** : Décision adoptée et en cours d’implémentation.
- **Contexte** :  
  Avec l’adoption d’une architecture microservices, il est nécessaire de définir un protocole d’échange de données entre les différents services. Deux options principales ont été considérées : HTTP/REST (avec JSON) et Protobuf/gRPC. HTTP/REST est largement adopté, bien documenté, et bénéficie d’un vaste écosystème d’outils pour le développement, le monitoring et le débogage. Protobuf/gRPC offre de meilleures performances et une sérialisation plus efficace, mais au prix d’une courbe d’apprentissage plus élevée et d’une interopérabilité réduite avec certains outils ou langages.
- **Décision** :  
  Pour la communication entre les microservices, le protocole HTTP/REST avec le format JSON sera utilisé. Ce choix permet de standardiser les échanges, de faciliter l’intégration avec des outils tiers, et de simplifier le développement et le débogage grâce à la lisibilité des messages.
- **Conséquences** :  
  - **Positives** :
    - Simplifie l’intégration avec des outils de monitoring, de logging et de test grâce à l’utilisation de HTTP/REST.
    - Rend le système plus accessible et interopérable avec d’autres systèmes ou langages grâce à l’utilisation de standards ouverts.
    - Facilite le développement et le débogage grâce à la lisibilité des messages JSON.
    - Permet une adoption rapide par l’équipe grâce à la familiarité avec HTTP/REST.
  - **Négatives** :
    - Les performances de HTTP/REST sont inférieures à celles de Protobuf/gRPC, ce qui peut être un facteur limitant pour des échanges très volumineux ou très fréquents.
    - La sérialisation JSON est moins efficace que Protobuf en termes de taille et de rapidité.
    - Nécessite une gestion manuelle de la documentation des API (bien que des outils comme Swagger puissent aider).
- **Compliance** :  
  Cette décision est cohérente avec les besoins actuels du projet, la maturité de l’équipe, et l’écosystème technologique choisi. Elle n’empêche pas une évolution future vers d’autres protocoles (comme gRPC) si les besoins de performance deviennent prioritaires.