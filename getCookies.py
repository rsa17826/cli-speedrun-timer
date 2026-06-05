import time
from evdev import UInput, ecodes as e

def move_mouse_absolute(x, y):
  # Create a virtual mouse capabilities layout
  cap = {
    e.EV_REL: [e.REL_X, e.REL_Y],
    e.EV_KEY: [e.BTN_LEFT, e.BTN_RIGHT]
  }

  # Initialize the virtual device
  with UInput(cap, name='virtual-mouse') as ui:
    # Give the OS half a second to register the new virtual device
    time.sleep(0.5)

    # Step 1: Force the cursor to (0, 0) by moving it by an extreme negative amount
    ui.write(e.EV_REL, e.REL_X, -10000)
    ui.write(e.EV_REL, e.REL_Y, -10000)
    ui.syn()

    # Step 2: Move relatively from (0,0) to your target destination
    ui.write(e.EV_REL, e.REL_X, x)
    ui.write(e.EV_REL, e.REL_Y, y)
    ui.syn()

if __name__ == "__main__":
  print("Moving mouse to (150, 150)...")
  move_mouse_absolute(150, 150)