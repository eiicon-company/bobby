"""
Test file to validate that Bobo properly extends Bob
"""

from bob import Bob
from bobo import Bobo


def test_bob_basic_functionality():
    """Test Bob's basic functionality"""
    bob = Bob()
    
    # Test basic properties
    assert bob.name == "Bob"
    assert "Hello, I'm Bob!" in bob.greet()
    assert "My name is Bob." in bob.introduce()
    
    # Test with custom name
    custom_bob = Bob("Robert")
    assert custom_bob.name == "Robert"
    assert "Hello, I'm Robert!" in custom_bob.greet()


def test_bobo_extends_bob():
    """Test that Bobo properly inherits from Bob"""
    bobo = Bobo()
    
    # Test inheritance - Bobo should be an instance of Bob
    assert isinstance(bobo, Bob)
    assert isinstance(bobo, Bobo)
    
    # Test that inherited methods work
    assert bobo.name == "Bobo"
    assert "My name is Bobo." in bobo.introduce()  # This should work via inheritance
    
    # Test enhanced greeting
    greeting = bobo.greet()
    assert "Hello, I'm Bobo!" in greeting
    assert "Nice to meet you" in greeting
    assert "!!!!!" in greeting  # Default enthusiasm level 5


def test_bobo_enhanced_functionality():
    """Test Bobo's additional functionality"""
    bobo = Bobo("Bobby", enthusiasm_level=3)
    
    # Test enhanced functionality
    assert bobo.enthusiasm_level == 3
    assert "dancing" in bobo.dance().lower()
    
    # Test enthusiasm adjustment
    result = bobo.set_enthusiasm(8)
    assert bobo.enthusiasm_level == 8
    assert "level 8" in result
    
    # Test enthusiasm bounds
    bobo.set_enthusiasm(15)  # Should cap at 10
    assert bobo.enthusiasm_level == 10
    
    bobo.set_enthusiasm(-5)  # Should floor at 0
    assert bobo.enthusiasm_level == 0


def test_method_overriding():
    """Test that Bobo properly overrides Bob's methods"""
    bob = Bob()
    bobo = Bobo()
    
    # Both should have greet method, but Bobo's should be enhanced
    bob_greeting = bob.greet()
    bobo_greeting = bobo.greet()
    
    assert bob_greeting != bobo_greeting  # Should be different
    assert "Nice to meet you" not in bob_greeting  # Bob's version is simpler
    assert "Nice to meet you" in bobo_greeting  # Bobo's version is enhanced


if __name__ == "__main__":
    print("Running inheritance tests...")
    
    test_bob_basic_functionality()
    print("✓ Bob basic functionality tests passed")
    
    test_bobo_extends_bob()
    print("✓ Bobo inheritance tests passed")
    
    test_bobo_enhanced_functionality()
    print("✓ Bobo enhanced functionality tests passed")
    
    test_method_overriding()
    print("✓ Method overriding tests passed")
    
    print("\nAll tests passed! Bobo successfully extends Bob.")