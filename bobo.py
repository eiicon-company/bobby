"""
Bobo - Extended class that inherits from Bob
"""

from bob import Bob


class Bobo(Bob):
    """Extended class that inherits from Bob and adds additional functionality"""
    
    def __init__(self, name="Bobo", enthusiasm_level=5):
        super().__init__(name)
        self.enthusiasm_level = enthusiasm_level
    
    def greet(self):
        """Enhanced greeting with enthusiasm"""
        base_greeting = super().greet()
        excitement = "!" * self.enthusiasm_level
        return f"{base_greeting} Nice to meet you{excitement}"
    
    def dance(self):
        """Bobo-specific functionality - dancing"""
        moves = ["spin", "jump", "wiggle"]
        return f"{self.name} is dancing: {' -> '.join(moves)}!"
    
    def set_enthusiasm(self, level):
        """Set the enthusiasm level for interactions"""
        self.enthusiasm_level = max(0, min(10, level))  # Keep between 0-10
        return f"{self.name}'s enthusiasm is now at level {self.enthusiasm_level}!"
    
    def __str__(self):
        return f"Bobo(name='{self.name}', enthusiasm_level={self.enthusiasm_level})"