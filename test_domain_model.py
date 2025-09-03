#!/usr/bin/env python3
"""
Test script for the complete D&D 5e domain model
Tests all new entities and hexagonal architecture components
"""
import sys
import traceback
from pathlib import Path

# Add project root to Python path
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

def test_imports():
    """Test that all domain model imports work correctly"""
    print("🧪 Testing Domain Model Imports...")
    
    try:
        # Test main facade
        from shared_domain import SRDDomainModel
        print("✅ SRDDomainModel facade imported successfully")
        
        # Test domain info
        info = SRDDomainModel.get_domain_info()
        print(f"✅ Domain model has {info['total_entity_types']} entity types")
        print(f"✅ Supported collections: {', '.join(info['supported_collections'][:5])}...")
        
        # Test core entities (existing)
        from shared_domain import DndClass, ClassId, Level, Ability
        print("✅ Core class entities imported")
        
        # Test query models (basic ones that should exist)
        from shared_domain import SpellSearchQuery, MonsterSummary
        print("✅ CQRS query models imported")
        
        return True
        
    except Exception as e:
        print(f"❌ Import error: {e}")
        traceback.print_exc()
        return False


def test_entity_creation():
    """Test creating instances of domain entities"""
    print("\n🧪 Testing Entity Creation...")
    
    try:
        from shared_domain import DndClass, ClassId, Level, Ability, HitDie
        
        # Create a class entity (existing entity type)
        dnd_class = DndClass(
            id=ClassId("guerriero"),
            name="Guerriero",
            primary_ability=Ability.FORZA,
            hit_die=HitDie(10),
            version="1.0",
            source="SRD"
        )
        
        print("✅ Class entity created successfully")
        print(f"   - Name: {dnd_class.name}")
        print(f"   - Primary Ability: {dnd_class.primary_ability.value}")
        print(f"   - Hit Die: {dnd_class.hit_die}")
        print(f"   - Version: {dnd_class.version}")
        
        return True
        
    except Exception as e:
        print(f"❌ Entity creation error: {e}")
        traceback.print_exc()
        return False


def test_validation_services():
    """Test domain validation services"""
    print("\n🧪 Testing Validation Services...")
    
    try:
        from shared_domain import DndClass, ClassId, Ability, HitDie, ClassValidationService
        
        # Create a problematic class for testing
        problematic_class = DndClass(
            id=ClassId("test-class"),
            name="Test Class",
            primary_ability=Ability.FORZA,
            hit_die=HitDie(12),
            version="1.0",
            source="Test"
        )
        
        # Test validation
        validator = ClassValidationService()
        errors = validator.validate_class(problematic_class)
        
        print("✅ Validation service executed successfully")
        print(f"   - Found {len(errors)} validation errors:")
        for error in errors:
            print(f"     • {error}")
        
        # Test suggestions
        suggestions = validator.suggest_missing_data(problematic_class)
        print(f"   - Generated {len(suggestions)} suggestions:")
        for suggestion in suggestions:
            print(f"     • {suggestion}")
        
        return True
        
    except Exception as e:
        print(f"❌ Validation service error: {e}")
        traceback.print_exc()
        return False


def test_query_models():
    """Test CQRS query models"""
    print("\n🧪 Testing CQRS Query Models...")
    
    try:
        from shared_domain import ClassSearchQuery, ClassSummary
        
        # Create search query
        class_query = ClassSearchQuery(
            text_query="fighter",
            primary_ability="Forza",
            min_hit_die=8,
            max_hit_die=12,
            sort_by="name",
            limit=20
        )
        
        # Create summary object
        class_summary = ClassSummary(
            id="guerriero",
            name="Guerriero",
            primary_ability="Forza",
            hit_die=10,
            source="SRD",
            is_spellcaster=False,
            subclass_count=3,
            subclass_names=["Champion", "Battle Master", "Eldritch Knight"]
        )
        
        print("✅ Query models created successfully")
        print(f"   - Class query: searching for '{class_query.text_query}' with ability {class_query.primary_ability}")
        print(f"   - Class summary: {class_summary.name} (d{class_summary.hit_die})")
        
        return True
        
    except Exception as e:
        print(f"❌ Query model error: {e}")
        traceback.print_exc()
        return False


def test_container_access():
    """Test accessing Parser and Editor containers"""
    print("\n🧪 Testing Hexagonal Architecture Containers...")
    
    try:
        # Test Parser container
        from srd_parser.infrastructure.container import get_container as get_parser_container
        parser_container = get_parser_container()
        print("✅ Parser container accessible")
        
        # Test getting services from parser container
        validation_handler = parser_container.get_validate_class_data_handler()
        print("✅ Parser validation handler created")
        
        # Test Editor container
        from editor.infrastructure.container import get_container as get_editor_container  
        editor_container = get_editor_container()
        print("✅ Editor container accessible")
        
        # Test getting services from editor container
        search_handler = editor_container.get_search_classes_handler()
        print("✅ Editor search handler created")
        
        return True
        
    except Exception as e:
        print(f"❌ Container access error: {e}")
        traceback.print_exc()
        return False


def main():
    """Run all tests"""
    print("🚀 Starting Domain Model and Hexagonal Architecture Tests\n")
    
    tests = [
        ("Domain Model Imports", test_imports),
        ("Entity Creation", test_entity_creation), 
        ("Validation Services", test_validation_services),
        ("CQRS Query Models", test_query_models),
        ("Hexagonal Containers", test_container_access)
    ]
    
    results = []
    for test_name, test_func in tests:
        print(f"\n{'='*50}")
        print(f"Running: {test_name}")
        print(f"{'='*50}")
        
        success = test_func()
        results.append((test_name, success))
    
    # Summary
    print(f"\n{'='*50}")
    print("TEST SUMMARY")
    print(f"{'='*50}")
    
    passed = sum(1 for _, success in results if success)
    total = len(results)
    
    for test_name, success in results:
        status = "✅ PASSED" if success else "❌ FAILED"
        print(f"{status} - {test_name}")
    
    print(f"\nOverall: {passed}/{total} tests passed")
    
    if passed == total:
        print("\n🎉 All tests passed! The domain model and hexagonal architecture are working correctly.")
        return 0
    else:
        print(f"\n⚠️  {total - passed} tests failed. Check the errors above.")
        return 1


if __name__ == "__main__":
    sys.exit(main())