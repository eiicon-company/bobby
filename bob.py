"""
Bob - Base class with fundamental functionality
"""


class Bob:
    """Base class that provides fundamental behavior"""
    
    def __init__(self, name="Bob"):
        self.name = name
    
    def greet(self):
        """Basic greeting functionality"""
        return f"Hello, I'm {self.name}!"
    
    def introduce(self):
        """Basic introduction"""
        return f"My name is {self.name}."
    
    def __str__(self):
        return f"Bob(name='{self.name}')"
    
    def __repr__(self):
        return self.__str__()