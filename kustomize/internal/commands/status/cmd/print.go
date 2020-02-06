// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io"
	"time"

	"sigs.k8s.io/kustomize/kstatus/status"
	"sigs.k8s.io/kustomize/kstatus/wait"
)

const (
	typeColumn      = "type"
	namespaceColumn = "namespace"
	nameColumn      = "name"
	statusColumn    = "status"
	messageColumn   = "message"

	RESET         = 0
	ESC           = 27
	RED     color = 31
	GREEN   color = 32
	YELLOW  color = 33
	DEFAULT color = -1 // This is not a valid ANSI escape code. It is used here to mean that no color should be set.
)

type color int

func moveUp(w io.Writer, lineCount int) {
	printOrDie(w, "%c[%dA", ESC, lineCount)
}

func eraseCurrentLine(w io.Writer) {
	printOrDie(w, "%c[2K\r", ESC)
}

type colorFunc func(s status.Status) color
type contentFunc func(resource ResourceStatusData) string

type tableColumnInfo struct {
	header      string
	width       int
	colorFunc   colorFunc
	contentFunc contentFunc
}

func defaultColorFunc(_ status.Status) color {
	return DEFAULT
}

var (
	tableColumns = map[string]tableColumnInfo{
		typeColumn: {
			header:    "TYPE",
			width:     25,
			colorFunc: defaultColorFunc,
			contentFunc: func(data ResourceStatusData) string {
				return fmt.Sprintf("%s/%s", data.Identifier.GroupKind.Group, data.Identifier.GroupKind.Kind)
			},
		},
		namespaceColumn: {
			header:    "NAMESPACE",
			width:     15,
			colorFunc: defaultColorFunc,
			contentFunc: func(data ResourceStatusData) string {
				return data.Identifier.Namespace
			},
		},
		nameColumn: {
			header:    "NAME",
			width:     20,
			colorFunc: defaultColorFunc,
			contentFunc: func(data ResourceStatusData) string {
				return data.Identifier.Name
			},
		},
		statusColumn: {
			header:    "STATUS",
			width:     10,
			colorFunc: colorForStatus,
			contentFunc: func(data ResourceStatusData) string {
				return data.Status.String()
			},
		},
		messageColumn: {
			header:    "MESSAGE",
			width:     40,
			colorFunc: defaultColorFunc,
			contentFunc: func(data ResourceStatusData) string {
				return data.Message
			},
		},
	}
	tableColumnOrder = []string{typeColumn, namespaceColumn, nameColumn, statusColumn, messageColumn}
)

type StatusInfo interface {
	CurrentStatus() StatusData
}

type StatusData struct {
	AggregateStatus  status.Status
	ResourceStatuses []ResourceStatusData
}

type ResourceStatusData struct {
	Identifier wait.ResourceIdentifier
	Status     status.Status
	Message    string
}

type TablePrinter struct {
	statusInfo    StatusInfo
	out           io.Writer
	err           io.Writer
	showAggStatus bool
}

func newTablePrinter(statusInfo StatusInfo, out io.Writer, err io.Writer, showAggStatus bool) *TablePrinter {
	return &TablePrinter{
		statusInfo:    statusInfo,
		out:           out,
		err:           err,
		showAggStatus: showAggStatus,
	}
}

func (s *TablePrinter) Print() {
	s.printTable(s.statusInfo.CurrentStatus(), false)
}

func (s *TablePrinter) PrintUntil(stop <-chan struct{}, interval time.Duration) <-chan struct{} {
	completed := make(chan struct{})
	s.printTable(s.statusInfo.CurrentStatus(), false)
	go func() {
		defer close(completed)
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-stop:
				ticker.Stop()
				s.printTable(s.statusInfo.CurrentStatus(), true)
				return
			case <-ticker.C:
				s.printTable(s.statusInfo.CurrentStatus(), true)
			}
		}
	}()
	return completed
}

func (s *TablePrinter) printTable(data StatusData, deleteUp bool) {
	if deleteUp {
		if s.showAggStatus {
			moveUp(s.out, 1)
		}
		moveUp(s.out, 1)
		moveUp(s.out, len(data.ResourceStatuses))
	}
	eraseCurrentLine(s.out)
	if s.showAggStatus {
		printOrDie(s.out, "AggregateStatus: ")
		printWithColorOrDie(s.out, colorForStatus(data.AggregateStatus), "%s\n", data.AggregateStatus)
	}
	s.printTableRow(headers())
	for _, resource := range data.ResourceStatuses {
		s.printTableRow(row(resource))
	}
}

func (s *TablePrinter) printTableRow(rowData []RowData) {
	for i, row := range rowData {

		format := fmt.Sprintf("%%-%ds", row.width)
		printWithColorOrDie(s.out, row.color, format, trimString(row.content, row.width))
		if i != len(rowData)-1 {
			printOrDie(s.out, "  ")
		}
	}
	printOrDie(s.out, "\n")
}

type RowData struct {
	content string
	color   color
	width   int
}

func headers() []RowData {
	var headers []RowData
	for _, columnName := range tableColumnOrder {
		column := tableColumns[columnName]
		headers = append(headers, RowData{
			content: column.header,
			color:   DEFAULT,
			width:   column.width,
		})
	}
	return headers
}

func row(resource ResourceStatusData) []RowData {
	var row []RowData
	for _, columnName := range tableColumnOrder {
		column := tableColumns[columnName]
		row = append(row, RowData{
			content: column.contentFunc(resource),
			color:   column.colorFunc(resource.Status),
			width:   column.width,
		})
	}
	return row
}

type eventContentFunc func(wait.Event) string

type eventColumnInfo struct {
	header                     string
	width                      int
	requireResourceUpdateEvent bool
	contentFunc                eventContentFunc
}

var (
	eventColumns = []eventColumnInfo{
		{
			header:                     "NAMESPACE",
			width:                      15,
			requireResourceUpdateEvent: true,
			contentFunc: func(event wait.Event) string {
				return event.EventResource.ResourceIdentifier.Namespace
			},
		},
		{
			header:                     "AGG STATUS",
			width:                      10,
			requireResourceUpdateEvent: false,
			contentFunc: func(event wait.Event) string {
				return event.AggregateStatus.String()
			},
		},
		{
			header:                     "TYPE",
			width:                      20,
			requireResourceUpdateEvent: true,
			contentFunc: func(event wait.Event) string {
				return event.EventResource.ResourceIdentifier.GroupKind.Kind
			},
		},
		{
			header:                     "NAME",
			width:                      20,
			requireResourceUpdateEvent: true,
			contentFunc: func(event wait.Event) string {
				return event.EventResource.ResourceIdentifier.Name
			},
		},
		{
			header:                     "STATUS",
			width:                      10,
			requireResourceUpdateEvent: true,
			contentFunc: func(event wait.Event) string {
				return event.EventResource.Status.String()
			},
		},
		{
			header:                     "MESSAGE",
			width:                      50,
			requireResourceUpdateEvent: false,
			contentFunc: func(event wait.Event) string {
				switch event.Type {
				case wait.ResourceUpdate:
					if event.EventResource.Error != nil {
						return event.EventResource.Error.Error()
					}
					return event.EventResource.Message
				case wait.Aborted:
					return fmt.Sprint("Operation aborted before all resources have become Current")
				case wait.Completed:
					return fmt.Sprint("All resources have become Current")
				}
				return ""
			},
		},
	}
)

type EventPrinter struct {
	out io.Writer
	err io.Writer
}

func newEventPrinter(out io.Writer, err io.Writer) *EventPrinter {
	for _, column := range eventColumns {
		format := fmt.Sprintf("%%-%ds  ", column.width)
		printOrDie(out, format, column.header)
	}
	printOrDie(out, "\n")
	return &EventPrinter{
		out: out,
		err: err,
	}
}

func (e *EventPrinter) printEvent(event wait.Event) {
	for _, column := range eventColumns {
		var text string
		if event.Type != wait.ResourceUpdate && column.requireResourceUpdateEvent {
			text = ""
		} else {
			text = trimString(column.contentFunc(event), column.width)
		}
		format := fmt.Sprintf("%%-%ds  ", column.width)
		printOrDie(e.out, format, text)
	}
	printOrDie(e.out, "\n")
}

func printOrDie(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		panic(err)
	}
}

func printWithColorOrDie(w io.Writer, color color, format string, a ...interface{}) {
	if color == DEFAULT {
		printOrDie(w, format, a...)
	} else {
		printOrDie(w, "%c[%dm", ESC, color)
		printOrDie(w, format, a...)
		printOrDie(w, "%c[%dm", ESC, RESET)
	}
}

func colorForStatus(s status.Status) color {
	switch s {
	case status.CurrentStatus:
		return GREEN
	case status.InProgressStatus:
		return YELLOW
	case status.FailedStatus:
		return RED
	}
	return DEFAULT
}

func trimString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}
	return str[:maxLength]
}
