"""
Demonstrate that Bobo extends Bob
"""

from bob import Bob
from bobo import Bobo


def main():
    print("=== Demonstrating: Bobo extends Bob ===\n")
    
    # Create instances
    print("1. Creating instances:")
    bob = Bob()
    bobo = Bobo()
    print(f"   Bob: {bob}")
    print(f"   Bobo: {bobo}")
    print()
    
    # Show inheritance relationship
    print("2. Inheritance relationship:")
    print(f"   isinstance(bobo, Bob): {isinstance(bobo, Bob)}")
    print(f"   isinstance(bobo, Bobo): {isinstance(bobo, Bobo)}")
    print(f"   isinstance(bob, Bobo): {isinstance(bob, Bobo)}")
    print()
    
    # Compare basic functionality
    print("3. Comparing basic functionality:")
    print(f"   Bob.greet(): {bob.greet()}")
    print(f"   Bobo.greet(): {bobo.greet()}")
    print()
    
    # Show inherited methods
    print("4. Inherited methods work in Bobo:")
    print(f"   Bob.introduce(): {bob.introduce()}")
    print(f"   Bobo.introduce(): {bobo.introduce()}")
    print()
    
    # Show Bobo's enhanced functionality
    print("5. Bobo's enhanced functionality:")
    print(f"   Bobo.dance(): {bobo.dance()}")
    print(f"   Bobo.set_enthusiasm(8): {bobo.set_enthusiasm(8)}")
    print(f"   Bobo.greet() after enthusiasm change: {bobo.greet()}")
    print()
    
    # Show customization
    print("6. Customization examples:")
    custom_bob = Bob("Robert")
    custom_bobo = Bobo("Bobby", enthusiasm_level=2)
    print(f"   Custom Bob: {custom_bob}")
    print(f"   Custom Bob greet: {custom_bob.greet()}")
    print(f"   Custom Bobo: {custom_bobo}")
    print(f"   Custom Bobo greet: {custom_bobo.greet()}")
    print()
    
    print("✨ Demonstration complete: Bobo successfully extends Bob! ✨")


if __name__ == "__main__":
    main()