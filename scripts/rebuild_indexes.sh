#!/bin/bash
# Rebuild all MongoDB text indexes

set -e

echo "Connecting to MongoDB and dropping/recreating text indexes..."

docker compose exec -T mongo mongosh --username admin --password password --authenticationDatabase admin <<EOF
use dnd

// Collections to update
const collections = [
  'incantesimi', 'mostri', 'classi', 'backgrounds', 'equipaggiamenti',
  'oggetti_magici', 'armi', 'armature', 'talenti', 'servizi',
  'strumenti', 'animali', 'regole', 'cavalcature_veicoli'
];

collections.forEach(collectionName => {
  const coll = db.getCollection(collectionName);
  
  // List and drop existing text indexes
  const indexes = coll.getIndexes();
  indexes.forEach(idx => {
    if (idx.key.title === 'text' || idx.name.includes('text_search')) {
      console.log(\`Dropping index: \${idx.name} from \${collectionName}\`);
      coll.dropIndex(idx.name);
    }
  });
});

console.log('Dropped all existing text indexes');
EOF

echo "Indexes dropped. Restart the application to recreate them with the new structure."
