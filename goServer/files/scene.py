from manim import *

class SceneName(Scene):
    def construct(self):
        circle = Circle(radius=2, color=BLUE)
        square = Square(side_length=4, color=RED)
        self.play(Create(circle))
        self.play(Transform(circle, square), circle.animate.set_color(RED), rate_func=linear, run_time=3)
        self.wait()