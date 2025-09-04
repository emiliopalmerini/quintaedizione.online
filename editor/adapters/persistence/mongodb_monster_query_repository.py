"""
MongoDB implementation for monster queries optimized for read operations (CQRS Query side)
Editor service focuses on fast, complex queries with projections
"""
from typing import Dict, List, Optional, Any
import logging
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorCollection

from shared_domain.monster_entities import Monster, MonsterId, MonsterQueryRepository
from shared_domain.query_models import MonsterSearchQuery, MonsterSummary, MonsterDetail

logger = logging.getLogger(__name__)


class MongoDBMonsterQueryRepository(MonsterQueryRepository):
    """MongoDB implementation optimized for monster read operations"""
    
    def __init__(self, connection_string: str, database_name: str):
        self.client = AsyncIOMotorClient(connection_string)
        self.db = self.client[database_name]
        self.collection: AsyncIOMotorCollection = self.db.mostri
        
        # Ensure read-optimized indexes
        self._ensure_read_indexes()
    
    async def _ensure_read_indexes(self) -> None:
        """Create indexes optimized for read operations"""
        try:
            # Text search index for name and description
            await self.collection.create_index([
                ("name", "text"),
                ("nome", "text"),
                ("description", "text"),
                ("descrizione", "text")
            ])
            
            # Compound indexes for filtering
            await self.collection.create_index([
                ("tipo", 1),
                ("gs", 1),
                ("allineamento", 1)
            ])
            
            # Size and environment indexes
            await self.collection.create_index("taglia")
            await self.collection.create_index("ambiente")
            
            logger.info("Monster read-optimized indexes ensured")
            
        except Exception as e:
            logger.warning(f"Could not create monster indexes: {e}")
    
    async def find_by_id(self, monster_id: MonsterId) -> Optional[Monster]:
        """Find monster by ID with full details"""
        try:
            doc = await self.collection.find_one({"_id": monster_id.value})
            if not doc:
                return None
            return self._document_to_entity(doc)
        except Exception as e:
            logger.error(f"Error finding monster by ID {monster_id.value}: {e}")
            return None
    
    async def search_monsters(self, query: MonsterSearchQuery) -> List[MonsterSummary]:
        """Search monsters with filtering and return summaries"""
        try:
            mongo_query = self._build_search_query(query)
            
            # Use projection for performance - only summary fields
            projection = {
                "_id": 1,
                "name": 1,
                "nome": 1,
                "tipo": 1,
                "taglia": 1,
                "gs": 1,
                "allineamento": 1,
                "ca": 1,
                "pf": 1,
                "velocita": 1,
                "ambiente": 1
            }
            
            # Apply sorting and limits
            cursor = self.collection.find(mongo_query, projection)
            
            if query.sort_by == "name":
                cursor = cursor.sort([("nome", 1), ("name", 1)])
            elif query.sort_by == "challenge_rating":
                cursor = cursor.sort("gs", 1)
            elif query.sort_by == "type":
                cursor = cursor.sort("tipo", 1)
            elif query.sort_by == "size":
                cursor = cursor.sort("taglia", 1)
            
            if query.limit:
                cursor = cursor.limit(query.limit)
            if query.offset:
                cursor = cursor.skip(query.offset)
            
            docs = await cursor.to_list(length=None)
            return [self._document_to_summary(doc) for doc in docs]
            
        except Exception as e:
            logger.error(f"Error in monster search: {e}")
            return []
    
    async def get_monsters_by_type(self, creature_type: str) -> List[MonsterSummary]:
        """Get all monsters of specific type"""
        try:
            docs = await self.collection.find(
                {"tipo": creature_type},
                {
                    "_id": 1, "name": 1, "nome": 1, "tipo": 1, 
                    "taglia": 1, "gs": 1, "allineamento": 1,
                    "ca": 1, "pf": 1, "velocita": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting monsters by type {creature_type}: {e}")
            return []
    
    async def get_monsters_by_challenge_rating(self, min_cr: float, max_cr: float) -> List[MonsterSummary]:
        """Get monsters within challenge rating range"""
        try:
            docs = await self.collection.find(
                {"gs": {"$gte": min_cr, "$lte": max_cr}},
                {
                    "_id": 1, "name": 1, "nome": 1, "tipo": 1,
                    "taglia": 1, "gs": 1, "allineamento": 1,
                    "ca": 1, "pf": 1, "velocita": 1
                }
            ).sort("gs", 1).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting monsters by CR {min_cr}-{max_cr}: {e}")
            return []
    
    async def get_monsters_by_environment(self, environment: str) -> List[MonsterSummary]:
        """Get monsters found in specific environment"""
        try:
            docs = await self.collection.find(
                {"ambiente": environment},
                {
                    "_id": 1, "name": 1, "nome": 1, "tipo": 1,
                    "taglia": 1, "gs": 1, "allineamento": 1,
                    "ca": 1, "pf": 1, "ambiente": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting monsters by environment {environment}: {e}")
            return []
    
    # Implement missing abstract methods
    async def get_monster_detail(self, monster_id: MonsterId) -> Optional[MonsterDetail]:
        """Get detailed monster information by ID"""
        try:
            doc = await self.collection.find_one({"_id": monster_id.value})
            if not doc:
                return None
            return self._document_to_detail(doc)
        except Exception as e:
            logger.error(f"Error getting monster detail for {monster_id.value}: {e}")
            return None
    
    async def get_monsters_by_cr(self, challenge_rating: str) -> List[MonsterSummary]:
        """Get monsters by specific challenge rating"""
        try:
            # Handle both numeric and fraction CR formats
            cr_query = challenge_rating
            if challenge_rating.replace(".", "").replace("/", "").isdigit():
                # Convert to float for numeric comparison if needed
                try:
                    cr_float = float(challenge_rating) if "." in challenge_rating else challenge_rating
                    if "/" in challenge_rating:
                        parts = challenge_rating.split("/")
                        cr_float = float(parts[0]) / float(parts[1])
                    cr_query = cr_float
                except ValueError:
                    pass
            
            docs = await self.collection.find(
                {"gs": cr_query},
                {
                    "_id": 1, "name": 1, "nome": 1, "tipo": 1,
                    "taglia": 1, "gs": 1, "allineamento": 1,
                    "ca": 1, "pf": 1, "velocita": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting monsters by CR {challenge_rating}: {e}")
            return []
    
    async def get_spellcasting_monsters(self) -> List[MonsterSummary]:
        """Get monsters that can cast spells"""
        try:
            docs = await self.collection.find(
                {
                    "$or": [
                        {"incantesimi_innati": {"$exists": True, "$ne": {}}},
                        {"incantesimi_preparati": {"$exists": True, "$ne": {}}},
                        {"caratteristica_incantatore": {"$exists": True}},
                        {"cd_incantesimo": {"$exists": True}},
                        {"azioni.nome": {"$regex": "incantesim", "$options": "i"}},
                        {"tratti.nome": {"$regex": "incantesim", "$options": "i"}}
                    ]
                },
                {
                    "_id": 1, "name": 1, "nome": 1, "tipo": 1,
                    "taglia": 1, "gs": 1, "allineamento": 1,
                    "ca": 1, "pf": 1, "velocita": 1
                }
            ).sort("gs", 1).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting spellcasting monsters: {e}")
            return []
    
    async def get_legendary_monsters(self) -> List[MonsterSummary]:
        """Get monsters with legendary actions"""
        try:
            docs = await self.collection.find(
                {
                    "$or": [
                        {"azioni_leggendarie": {"$exists": True, "$ne": []}},
                        {"legendary_actions": {"$exists": True, "$ne": []}},
                        {"azioni_del_covo": {"$exists": True, "$ne": []}},
                        {"lair_actions": {"$exists": True, "$ne": []}}
                    ]
                },
                {
                    "_id": 1, "name": 1, "nome": 1, "tipo": 1,
                    "taglia": 1, "gs": 1, "allineamento": 1,
                    "ca": 1, "pf": 1, "velocita": 1
                }
            ).sort("gs", -1).to_list(length=None)  # Sort by CR descending
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting legendary monsters: {e}")
            return []
    
    def _build_search_query(self, query: MonsterSearchQuery) -> Dict[str, Any]:
        """Build MongoDB query from search parameters"""
        mongo_query = {}
        
        # Text search
        if query.text_query:
            mongo_query["$or"] = [
                {"name": {"$regex": query.text_query, "$options": "i"}},
                {"nome": {"$regex": query.text_query, "$options": "i"}},
                {"description": {"$regex": query.text_query, "$options": "i"}},
                {"descrizione": {"$regex": query.text_query, "$options": "i"}}
            ]
        
        # Filters
        if query.monster_type:
            mongo_query["tipo"] = query.monster_type
        
        if query.size:
            mongo_query["taglia"] = query.size
        
        if query.alignment:
            mongo_query["allineamento"] = query.alignment
        
        if query.min_challenge_rating is not None:
            mongo_query["gs"] = {"$gte": query.min_challenge_rating}
        if query.max_challenge_rating is not None:
            if "gs" in mongo_query:
                mongo_query["gs"]["$lte"] = query.max_challenge_rating
            else:
                mongo_query["gs"] = {"$lte": query.max_challenge_rating}
        
        if query.environment:
            mongo_query["ambiente"] = query.environment
        
        if query.min_armor_class is not None:
            mongo_query["ca"] = {"$gte": query.min_armor_class}
        if query.max_armor_class is not None:
            if "ca" in mongo_query:
                mongo_query["ca"]["$lte"] = query.max_armor_class
            else:
                mongo_query["ca"] = {"$lte": query.max_armor_class}
        
        if query.min_hit_points is not None:
            mongo_query["pf"] = {"$gte": query.min_hit_points}
        if query.max_hit_points is not None:
            if "pf" in mongo_query:
                mongo_query["pf"]["$lte"] = query.max_hit_points
            else:
                mongo_query["pf"] = {"$lte": query.max_hit_points}
        
        return mongo_query
    
    def _document_to_summary(self, doc: Dict[str, Any]) -> MonsterSummary:
        """Convert MongoDB document to MonsterSummary"""
        return MonsterSummary(
            id=str(doc["_id"]),
            nome=doc.get("nome", doc.get("name", "")),
            size=doc.get("taglia", ""),
            monster_type=doc.get("tipo", ""),
            alignment=doc.get("allineamento", ""),
            challenge_rating=str(doc.get("gs", 0)),
            armor_class=doc.get("ca", 0),
            hit_points=str(doc.get("pf", 0)),
            is_spellcaster=bool(doc.get("incantesimi_innati") or doc.get("incantesimi_preparati")),
            has_legendary_actions=bool(doc.get("azioni_leggendarie") or doc.get("legendary_actions"))
        )
    
    def _document_to_detail(self, doc: Dict[str, Any]) -> MonsterDetail:
        """Convert MongoDB document to MonsterDetail"""
        # Extract abilities safely
        abilities = doc.get("caratteristiche", {})
        if isinstance(abilities, dict):
            ability_scores = {
                "forza": abilities.get("forza", 10),
                "destrezza": abilities.get("destrezza", 10),
                "costituzione": abilities.get("costituzione", 10),
                "intelligenza": abilities.get("intelligenza", 10),
                "saggezza": abilities.get("saggezza", 10),
                "carisma": abilities.get("carisma", 10)
            }
        else:
            ability_scores = {"forza": 10, "destrezza": 10, "costituzione": 10, "intelligenza": 10, "saggezza": 10, "carisma": 10}

        # Extract spellcasting info if present
        spellcasting_info = None
        if doc.get("incantesimi_innati") or doc.get("incantesimi_preparati"):
            spellcasting_info = {
                "innate_spells": doc.get("incantesimi_innati", {}),
                "prepared_spells": doc.get("incantesimi_preparati", {}),
                "spellcasting_ability": doc.get("caratteristica_incantatore"),
                "spell_save_dc": doc.get("cd_incantesimo"),
                "spell_attack_bonus": doc.get("bonus_attacco_incantesimo")
            }

        # Convert speed to string
        speed_dict = doc.get("velocita", {})
        if isinstance(speed_dict, dict):
            speed_parts = []
            if "camminare" in speed_dict:
                speed_parts.append(f"{speed_dict['camminare']} m")
            if "volare" in speed_dict:
                speed_parts.append(f"volo {speed_dict['volare']} m")
            if "nuotare" in speed_dict:
                speed_parts.append(f"nuoto {speed_dict['nuotare']} m")
            speed_str = ", ".join(speed_parts) if speed_parts else "9 m"
        else:
            speed_str = str(speed_dict) if speed_dict else "9 m"

        # Extract actions, traits, etc.
        actions = []
        for action in doc.get("azioni", []):
            if isinstance(action, dict):
                actions.append({
                    "nome": action.get("nome", ""),
                    "descrizione": action.get("descrizione", "")
                })

        traits = []
        for trait in doc.get("tratti", []):
            if isinstance(trait, dict):
                traits.append({
                    "nome": trait.get("nome", ""),
                    "descrizione": trait.get("descrizione", "")
                })

        legendary_actions = []
        for leg_action in doc.get("azioni_leggendarie", []):
            if isinstance(leg_action, dict):
                legendary_actions.append({
                    "nome": leg_action.get("nome", ""),
                    "descrizione": leg_action.get("descrizione", "")
                })

        return MonsterDetail(
            id=str(doc["_id"]),
            nome=doc.get("nome", doc.get("name", "")),
            size=doc.get("taglia", ""),
            monster_type=doc.get("tipo", ""),
            alignment=doc.get("allineamento", ""),
            armor_class=doc.get("ca", 10),
            hit_points=str(doc.get("pf", "1 (1d4)")),
            speed=speed_str,
            challenge_rating=str(doc.get("gs", 0)),
            xp_value=doc.get("xp", 0),
            abilities=ability_scores,
            saving_throws=doc.get("tiri_salvezza", {}),
            skills=doc.get("competenze", {}),
            damage_resistances=doc.get("resistenze_danni", []),
            damage_immunities=doc.get("immunita_danni", []),
            condition_immunities=doc.get("immunita_condizioni", []),
            senses=doc.get("sensi", []),
            languages=doc.get("linguaggi", []),
            traits=traits,
            actions=actions,
            legendary_actions=legendary_actions,
            is_spellcaster=bool(doc.get("incantesimi_innati") or doc.get("incantesimi_preparati")),
            spellcasting_info=spellcasting_info
        )
    
    def _document_to_entity(self, doc: Dict[str, Any]) -> Monster:
        """Convert MongoDB document to full domain entity"""
        # For now, use a simplified conversion
        # In production, this should match the parser's entity conversion
        from shared_domain.monster_entities import (
            CreatureType, Size, Alignment, ChallengeRating, AbilityScores
        )
        
        return Monster(
            id=MonsterId(str(doc["_id"])),
            name=doc.get("name", doc.get("nome", "")),
            italian_name=doc.get("nome", doc.get("name", "")),
            creature_type=CreatureType(doc.get("tipo", "")),
            size=Size(doc.get("taglia", "")),
            alignment=Alignment(doc.get("allineamento", "")),
            challenge_rating=ChallengeRating(doc.get("gs", 0)),
            armor_class=doc.get("ca", 0),
            hit_points=doc.get("pf", 0),
            speed=doc.get("velocita", {}),
            ability_scores=AbilityScores(
                strength=doc.get("caratteristiche", {}).get("forza", 10),
                dexterity=doc.get("caratteristiche", {}).get("destrezza", 10),
                constitution=doc.get("caratteristiche", {}).get("costituzione", 10),
                intelligence=doc.get("caratteristiche", {}).get("intelligenza", 10),
                wisdom=doc.get("caratteristiche", {}).get("saggezza", 10),
                charisma=doc.get("caratteristiche", {}).get("carisma", 10)
            ),
            skills=doc.get("competenze", {}),
            senses=doc.get("sensi", {}),
            languages=doc.get("linguaggi", []),
            actions=doc.get("azioni", []),
            legendary_actions=doc.get("azioni_leggendarie", []),
            description=doc.get("description", doc.get("descrizione", "")),
            source="SRD"
        )
    
    async def close(self) -> None:
        """Close database connection"""
        self.client.close()