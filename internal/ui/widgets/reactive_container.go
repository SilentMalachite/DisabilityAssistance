package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// ReactiveContainer is a container that automatically updates its content based on app state
type ReactiveContainer struct {
	appState       *AppState
	container      *fyne.Container
	contentFunc    func(*AppState) fyne.CanvasObject
	updateFunc     func()
	currentContent fyne.CanvasObject
}

// OnStateChanged implements StateObserver interface
func (rc *ReactiveContainer) OnStateChanged() {
	rc.Update()
}

// NewReactiveContainer creates a new reactive container
func NewReactiveContainer(appState *AppState, contentFunc func(*AppState) fyne.CanvasObject) *ReactiveContainer {
	rc := &ReactiveContainer{
		appState:    appState,
		contentFunc: contentFunc,
	}

	// Create initial content
	rc.currentContent = rc.contentFunc(rc.appState)

	// Create container with initial content
	rc.container = container.NewWithoutLayout(rc.currentContent)

	// Register as observer for automatic updates
	rc.appState.AddObserver(rc)

	return rc
}

// Update refreshes the container content based on current app state
func (rc *ReactiveContainer) Update() {
	// Generate new content
	newContent := rc.contentFunc(rc.appState)

	// Update container content
	rc.currentContent = newContent
	rc.container.RemoveAll()
	rc.container.Add(newContent)

	// Call update callback if set
	if rc.updateFunc != nil {
		rc.updateFunc()
	}

	// Refresh the container
	rc.container.Refresh()
}

// GetContent returns the current content
func (rc *ReactiveContainer) GetContent() fyne.CanvasObject {
	return rc.currentContent
}

// GetContainer returns the underlying container
func (rc *ReactiveContainer) GetContainer() *fyne.Container {
	return rc.container
}

// SetUpdateCallback sets a callback function to be called on updates
func (rc *ReactiveContainer) SetUpdateCallback(callback func()) {
	rc.updateFunc = callback
}

// Destroy removes this container as an observer and cleans up resources
func (rc *ReactiveContainer) Destroy() {
	rc.appState.RemoveObserver(rc)
}
