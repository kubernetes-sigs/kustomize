// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/sethgrid/curse"

	"sigs.k8s.io/kustomize/kstatus/status"
	"sigs.k8s.io/kustomize/kstatus/wait"
)

const (
	typeColumn      = "type"
	namespaceColumn = "namespace"
	nameColumn      = "name"
	statusColumn    = "status"
	messageColumn   = "message"
)

type colorFunc func(s status.Status) int
type contentFunc func(resource ResourceStatusData) string

type tableColumnInfo struct {
	header      string
	width       int
	colorFunc   colorFunc
	contentFunc contentFunc
}

func defaultColorFunc(_ status.Status) int {
	return curse.WHITE
}

var (
	tableColumns = map[string]tableColumnInfo{
		typeColumn: {
			header:    "TYPE",
			width:     25,
			colorFunc: defaultColorFunc,
			contentFunc: func(data ResourceStatusData) string {
				return fmt.Sprintf("%s/%s", data.Identifier.GetAPIVersion(),
					data.Identifier.GetKind())
			},
		},
		namespaceColumn: {
			header:    "NAMESPACE",
			width:     15,
			colorFunc: defaultColorFunc,
			contentFunc: func(data ResourceStatusData) string {
				return data.Identifier.GetNamespace()
			},
		},
		nameColumn: {
			header:    "NAME",
			width:     20,
			colorFunc: defaultColorFunc,
			contentFunc: func(data ResourceStatusData) string {
				return data.Identifier.GetName()
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
	c := newCurseOrDie()
	s.printTable(c, s.statusInfo.CurrentStatus(), false)
}

func (s *TablePrinter) PrintUntil(stop <-chan struct{}, interval time.Duration) <-chan struct{} {
	completed := make(chan struct{})
	go func() {
		defer close(completed)
		c := newCurseOrDie()
		c.SetDefaultStyle()
		s.printTable(c, s.statusInfo.CurrentStatus(), false)
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-stop:
				ticker.Stop()
				s.printTable(c, s.statusInfo.CurrentStatus(), true)
				return
			case <-ticker.C:
				s.printTable(c, s.statusInfo.CurrentStatus(), true)
			}
		}
	}()
	return completed
}

func (s *TablePrinter) printTable(c *curse.Cursor, data StatusData, moveUp bool) {
	if moveUp {
		if s.showAggStatus {
			c.MoveUp(1)
		}
		c.MoveUp(1)
		c.MoveUp(len(data.ResourceStatuses))
	}
	c.EraseCurrentLine()
	if s.showAggStatus {
		printOrDie(s.out, "AggregateStatus: ")
		c.SetColor(colorForStatus(data.AggregateStatus))
		printOrDie(s.out, "%s\n", data.AggregateStatus)
		c.SetDefaultStyle()
	}
	s.printTableRow(c, headers())
	for _, resource := range data.ResourceStatuses {
		s.printTableRow(c, row(resource))
	}
}

func (s *TablePrinter) printTableRow(c *curse.Cursor, rowData []RowData) {
	for _, row := range rowData {
		c.SetColor(row.color)
		format := fmt.Sprintf("%%-%ds  ", row.width)
		printOrDie(s.out, format, trimString(row.content, row.width))
		c.SetDefaultStyle()
	}
	printOrDie(s.out, "\n")
}

type RowData struct {
	content string
	color   int
	width   int
}

func headers() []RowData {
	var headers []RowData
	for _, columnName := range tableColumnOrder {
		column := tableColumns[columnName]
		headers = append(headers, RowData{
			content: column.header,
			color:   curse.WHITE,
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
			header:                     "EVENT TYPE",
			width:                      15,
			requireResourceUpdateEvent: false,
			contentFunc: func(event wait.Event) string {
				return string(event.Type)
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
				return fmt.Sprintf("%s/%s", event.EventResource.Identifier.GetAPIVersion(),
					event.EventResource.Identifier.GetKind())
			},
		},
		{
			header:                     "NAMESPACE",
			width:                      15,
			requireResourceUpdateEvent: true,
			contentFunc: func(event wait.Event) string {
				return event.EventResource.Identifier.GetNamespace()
			},
		},
		{
			header:                     "NAME",
			width:                      20,
			requireResourceUpdateEvent: true,
			contentFunc: func(event wait.Event) string {
				return event.EventResource.Identifier.GetName()
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
			requireResourceUpdateEvent: true,
			contentFunc: func(event wait.Event) string {
				if event.EventResource.Error != nil {
					return event.EventResource.Error.Error()
				}
				return event.EventResource.Message
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
		if event.Type != wait.ResourceUpdate && column.requireResourceUpdateEvent {
			continue
		}
		format := fmt.Sprintf("%%-%ds  ", column.width)
		printOrDie(e.out, format, trimString(column.contentFunc(event), column.width))
	}
	printOrDie(e.out, "\n")
}

func newCurseOrDie() *curse.Cursor {
	// TODO: Handle the issue with creating a new Cursor. For now we
	// are just ignoring the error (which mostly works).
	c, _ := curse.New()
	return c
}

func printOrDie(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		panic(err)
	}
}

func colorForStatus(s status.Status) int {
	switch s {
	case status.CurrentStatus:
		return curse.GREEN
	case status.UnknownStatus:
		return curse.WHITE
	case status.InProgressStatus:
		return curse.YELLOW
	case status.FailedStatus:
		return curse.RED
	}
	return curse.WHITE
}

func trimString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}
	return str[:maxLength]
}
