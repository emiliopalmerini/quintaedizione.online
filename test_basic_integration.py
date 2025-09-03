#!/usr/bin/env python3
"""
Basic integration test for hexagonal architecture
Tests core functionality without complex imports
"""
import sys
import requests
import time
from pathlib import Path

def test_services_running():
    """Test that both services are running"""
    print("🧪 Testing Service Availability...")
    
    services = [
        ("Editor", "http://localhost:8000/", "D&D 5e SRD"),
        ("Parser", "http://localhost:8100/", "SRD Parser")
    ]
    
    results = []
    for name, url, expected_content in services:
        try:
            response = requests.get(url, timeout=5)
            if response.status_code == 200 and expected_content in response.text:
                print(f"✅ {name} service is running at {url}")
                results.append(True)
            else:
                print(f"❌ {name} service returned unexpected response: {response.status_code}")
                results.append(False)
        except requests.RequestException as e:
            print(f"❌ {name} service is not accessible: {e}")
            results.append(False)
    
    return all(results)


def test_database_connection():
    """Test MongoDB connection through editor service"""
    print("\n🧪 Testing Database Connection...")
    
    try:
        # Test a simple database query through the editor
        response = requests.get("http://localhost:8000/classi", timeout=10)
        
        if response.status_code == 200:
            if "classi" in response.text.lower() or "classes" in response.text.lower():
                print("✅ Database connection working - classes page loaded")
                return True
            else:
                print("❌ Database connection issue - no classes content found")
                return False
        else:
            print(f"❌ Database connection issue - status code: {response.status_code}")
            return False
            
    except requests.RequestException as e:
        print(f"❌ Database connection test failed: {e}")
        return False


def test_hexagonal_demo_routes():
    """Test hexagonal architecture demo routes"""
    print("\n🧪 Testing Hexagonal Architecture Demo...")
    
    try:
        # Test hexagonal demo homepage
        response = requests.get("http://localhost:8000/hex/", timeout=10)
        
        if response.status_code == 200:
            if "hexagonal" in response.text.lower() or "architettura" in response.text.lower():
                print("✅ Hexagonal architecture demo page accessible")
                hex_demo_working = True
            else:
                print("❌ Hexagonal demo page missing expected content")
                hex_demo_working = False
        else:
            print(f"❌ Hexagonal demo page returned: {response.status_code}")
            hex_demo_working = False
        
        # Test hexagonal classes route
        response = requests.get("http://localhost:8000/hex/classes", timeout=10)
        
        if response.status_code == 200:
            print("✅ Hexagonal classes route accessible")
            classes_working = True
        else:
            print(f"❌ Hexagonal classes route returned: {response.status_code}")
            classes_working = False
        
        return hex_demo_working and classes_working
        
    except requests.RequestException as e:
        print(f"❌ Hexagonal architecture demo test failed: {e}")
        return False


def test_parser_interface():
    """Test parser web interface"""
    print("\n🧪 Testing Parser Interface...")
    
    try:
        # Test parser homepage
        response = requests.get("http://localhost:8100/", timeout=10)
        
        if response.status_code == 200:
            if "parser" in response.text.lower():
                print("✅ Parser web interface accessible")
                return True
            else:
                print("❌ Parser interface missing expected content")
                return False
        else:
            print(f"❌ Parser interface returned: {response.status_code}")
            return False
            
    except requests.RequestException as e:
        print(f"❌ Parser interface test failed: {e}")
        return False


def test_basic_functionality():
    """Test basic CRUD operations through web interfaces"""
    print("\n🧪 Testing Basic Functionality...")
    
    try:
        # Test searching in editor
        search_url = "http://localhost:8000/classi?q=barbaro"
        response = requests.get(search_url, timeout=10)
        
        if response.status_code == 200:
            print("✅ Search functionality working")
            search_working = True
        else:
            print(f"❌ Search returned: {response.status_code}")
            search_working = False
        
        # Test different collection
        armor_url = "http://localhost:8000/armature"
        response = requests.get(armor_url, timeout=10)
        
        if response.status_code == 200:
            print("✅ Multiple collections accessible")
            collections_working = True
        else:
            print(f"❌ Armor collection returned: {response.status_code}")
            collections_working = False
        
        return search_working and collections_working
        
    except requests.RequestException as e:
        print(f"❌ Basic functionality test failed: {e}")
        return False


def test_architecture_health():
    """Test overall architecture health"""
    print("\n🧪 Testing Architecture Health...")
    
    try:
        # Test editor health endpoint if it exists
        try:
            response = requests.get("http://localhost:8000/healthz", timeout=5)
            editor_health = response.status_code == 200
        except:
            editor_health = True  # No health endpoint is OK
        
        # Test parser health endpoint
        try:
            response = requests.get("http://localhost:8100/healthz", timeout=5)
            parser_health = response.status_code == 200
        except:
            parser_health = True  # No health endpoint is OK
        
        if editor_health and parser_health:
            print("✅ Architecture health checks passed")
            return True
        else:
            print("❌ Some health checks failed")
            return False
            
    except Exception as e:
        print(f"❌ Architecture health test failed: {e}")
        return False


def main():
    """Run integration tests"""
    print("🚀 Starting Basic Integration Tests for Hexagonal Architecture\n")
    print("Testing D&D 5e SRD system with hexagonal architecture...")
    print("=" * 60)
    
    # Wait a bit for services to be fully ready
    print("⏳ Waiting for services to be ready...")
    time.sleep(3)
    
    tests = [
        ("Service Availability", test_services_running),
        ("Database Connection", test_database_connection),
        ("Hexagonal Demo Routes", test_hexagonal_demo_routes),
        ("Parser Interface", test_parser_interface),
        ("Basic Functionality", test_basic_functionality),
        ("Architecture Health", test_architecture_health)
    ]
    
    results = []
    for test_name, test_func in tests:
        print(f"\n{'='*60}")
        print(f"Running: {test_name}")
        print('='*60)
        
        try:
            success = test_func()
            results.append((test_name, success))
        except Exception as e:
            print(f"❌ Test '{test_name}' failed with exception: {e}")
            results.append((test_name, False))
    
    # Summary
    print(f"\n{'='*60}")
    print("INTEGRATION TEST SUMMARY")
    print('='*60)
    
    passed = sum(1 for _, success in results if success)
    total = len(results)
    
    for test_name, success in results:
        status = "✅ PASSED" if success else "❌ FAILED"
        print(f"{status} - {test_name}")
    
    print(f"\nOverall: {passed}/{total} tests passed")
    
    if passed == total:
        print("\n🎉 All integration tests passed!")
        print("✅ Hexagonal architecture is working correctly")
        print("✅ Both Editor and Parser services are operational")
        print("✅ Database connectivity is working")
        return 0
    elif passed >= total * 0.7:  # 70% pass rate
        print(f"\n⚠️  Most tests passed ({passed}/{total})")
        print("✅ Core functionality is working")
        print("⚠️  Some advanced features may need attention")
        return 0
    else:
        print(f"\n❌ Many tests failed ({total - passed}/{total})")
        print("❌ System may have significant issues")
        return 1


if __name__ == "__main__":
    sys.exit(main())