package common

// Widget provides an interface to interact with widgets.
// A widget renders some sort of data to a termui widget.
type Widget interface {
	// Update is called by the application periodically and implementations must render their data
	// to a termui widget.
	Update()

	// Resize is called by the application when the terminal has been resized.
	Resize(int, int, int, int)

	// Close is called when the application exists. Widgets can implement this to return gracefully.
	Close()
}

// Scrollable is implemented by widgets which can scroll.
// Ideally, this is inherited from the termui List widget.
type Scrollable interface {
	ScrollUp()
	ScrollDown()
	ScrollPageUp()
	ScrollPageDown()
	ScrollTop()
	ScrollBottom()
}

type TextBox interface {
	Widget

	SetText(string)
}
