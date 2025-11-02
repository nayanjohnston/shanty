package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ContentModel struct {
	libraryModel tea.Model
	queueModel   tea.Model
}

func initContentModel(queue *Queue) ContentModel {
	return ContentModel{
		libraryModel: initLibraryModel(queue),
		queueModel:   initQueueModel(queue),
	}
}

func (m ContentModel) Init() tea.Cmd {
	return tea.Batch(
		m.libraryModel.Init(),
		m.queueModel.Init(),
	)
}

func (m ContentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			currentContentFocus = libraryFocus
		case "2":
			currentContentFocus = queueFocus
		}

		switch currentContentFocus {
		case libraryFocus:
			m.libraryModel, cmd = m.libraryModel.Update(msg)
		case queueFocus:
			m.queueModel, cmd = m.queueModel.Update(msg)
		}
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	m.libraryModel, cmd = m.libraryModel.Update(msg)
	cmds = append(cmds, cmd)

	m.queueModel, cmd = m.queueModel.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ContentModel) View() string {
	output := ""

	switch currentContentFocus {
	case libraryFocus:
		output = m.libraryModel.View()
	case queueFocus:
		output = m.queueModel.View()
	}

	return output
}
