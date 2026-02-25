Optimize the attached prompt for clarity and effectiveness. Output the refined version in a ~~~~ codeblock WITHOUT executing it. It should be optimized for the attached #codebase

Ensure the prompt:
- States a clear, actionable objective
- Specifies ONE execution mode: (1) autonomous action, (2) report/plan generation, or (3) interactive with user approval
- Uses precise language and logical structure
- Defines expected output format
- Includes success criteria when applicable

Preserve the original intent while eliminating ambiguity. If the execution mode is unclear, default to report generation and note this assumption.
Avoid excessively long responses which will exceed github copilot length limits for claude-sonnet-4.5

~~~~
The new combined onboarding tutorial and character customization system appears to be partly broken. Instead, a tutorial which is presented subtly out-of-order and skips steps is shown. Button presses advance the tutorial when we do not want them to, due to either control collision or sensitivity issues. We want the modern, fixed onboarding AND character creation tools exposed to new users. Present **all** steps in the correct, logical order. Carefully examine all tutorial systems, and ensure that the modern version with all specified features for tutorial guides and character development are included, enabled, added to the game, and bug-free. Players should be able to choose a starting class and equipment by completing the tutorial. If you discover bugs while integrating, fix them.

Character creation should start with the player being presented with a selection of random names and a default name. It should proceed to class selection, followed by subclass selection. Last, portrait generation should be presented.

Ensure all inputs are wired correctly for touch, keyboard, and game controller modes. Select, next, finish, and all options should function correctly. Assure that the layoud of the menus contains all menu items without causing them to overlap or be unreadable.
~~~~