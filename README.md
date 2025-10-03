# bobo

A Python implementation demonstrating inheritance where **Bobo extends Bob**.

## Overview

This repository contains a simple but complete implementation of class inheritance in Python:

- `Bob` - The base class with fundamental functionality
- `Bobo` - The extended class that inherits from Bob and adds enhanced features

## Classes

### Bob (Base Class)
- Basic greeting and introduction functionality
- Simple string representation
- Foundation for extension

### Bobo (Extended Class)
- **Extends Bob** through inheritance
- Enhanced greeting with enthusiasm levels
- Additional functionality like dancing
- Customizable enthusiasm settings

## Usage

```python
from bob import Bob
from bobo import Bobo

# Create instances
bob = Bob()
bobo = Bobo()

# Basic functionality (inherited)
print(bob.greet())        # "Hello, I'm Bob!"
print(bobo.greet())       # "Hello, I'm Bobo! Nice to meet you!!!!!"
print(bobo.introduce())   # "My name is Bobo." (inherited from Bob)

# Bobo's enhanced functionality
print(bobo.dance())       # "Bobo is dancing: spin -> jump -> wiggle!"
bobo.set_enthusiasm(8)    # Increase enthusiasm
print(bobo.greet())       # More exclamation marks!

# Inheritance verification
print(isinstance(bobo, Bob))   # True - Bobo extends Bob
print(isinstance(bobo, Bobo))  # True - Bobo is also a Bobo
```

## Running the Code

```bash
# Run tests to verify inheritance works correctly
python test_inheritance.py

# Run demonstration to see inheritance in action
python demo.py
```

## Key Features Demonstrated

1. **Class Inheritance**: Bobo inherits all methods and properties from Bob
2. **Method Overriding**: Bobo enhances Bob's `greet()` method while maintaining compatibility
3. **Super() Usage**: Proper use of `super()` to call parent class methods
4. **Additional Functionality**: Bobo adds new methods like `dance()` and `set_enthusiasm()`
5. **Polymorphism**: Bobo instances can be used wherever Bob instances are expected

This implementation satisfies the requirement that **"bobo extends bob"** using Python's class inheritance mechanism.