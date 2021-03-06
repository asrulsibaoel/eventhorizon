// Copyright (c) 2016 - Max Ekman <max@looplab.se>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"context"
	"strings"
	"testing"

	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/mocks"
)

// EventStoreCommonTests are test cases that are common to all implementations
// of event stores.
func EventStoreCommonTests(t *testing.T, ctx context.Context, store eh.EventStore) []eh.Event {
	savedEvents := []eh.Event{}

	ctx = context.WithValue(ctx, "testkey", "testval")

	t.Log("save no events")
	err := store.Save(ctx, []eh.Event{}, 0)
	if esErr, ok := err.(eh.EventStoreError); !ok || esErr.Err != eh.ErrNoEventsToAppend {
		t.Error("there shoud be a ErrNoEventsToAppend error:", err)
	}

	t.Log("save event, version 1")
	id, _ := eh.ParseUUID("c1138e5f-f6fb-4dd0-8e79-255c6c8d3756")
	agg := mocks.NewAggregate(id)
	event1 := agg.NewEvent(mocks.EventType, &mocks.EventData{"event1"})
	agg.ApplyEvent(ctx, event1) // Apply event to increment the aggregate version.
	err = store.Save(ctx, []eh.Event{event1}, 0)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	savedEvents = append(savedEvents, event1)
	if val, ok := agg.Context.Value("testkey").(string); !ok || val != "testval" {
		t.Error("the context should be correct:", agg.Context)
	}

	t.Log("try to save same event twice")
	err = store.Save(ctx, []eh.Event{event1}, 1)
	if esErr, ok := err.(eh.EventStoreError); !ok || esErr.Err != eh.ErrIncorrectEventVersion {
		t.Error("there should be a ErrIncerrectEventVersion error:", err)
	}

	t.Log("save event, version 2")
	event2 := agg.NewEvent(mocks.EventType, &mocks.EventData{"event2"})
	agg.ApplyEvent(ctx, event2) // Apply event to increment the aggregate version.
	err = store.Save(ctx, []eh.Event{event2}, 1)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	savedEvents = append(savedEvents, event2)

	t.Log("save event without data, version 3")
	event3 := agg.NewEvent(mocks.EventOtherType, nil)
	agg.ApplyEvent(ctx, event3) // Apply event to increment the aggregate version.
	err = store.Save(ctx, []eh.Event{event3}, 2)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	savedEvents = append(savedEvents, event3)

	t.Log("save multiple events, version 4, 5 and 6")
	event4 := agg.NewEvent(mocks.EventOtherType, nil)
	agg.ApplyEvent(ctx, event4) // Apply event to increment the aggregate version.
	event5 := agg.NewEvent(mocks.EventOtherType, nil)
	agg.ApplyEvent(ctx, event5) // Apply event to increment the aggregate version.
	event6 := agg.NewEvent(mocks.EventOtherType, nil)
	agg.ApplyEvent(ctx, event6) // Apply event to increment the aggregate version.
	err = store.Save(ctx, []eh.Event{event4, event5, event6}, 3)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	savedEvents = append(savedEvents, event4, event5, event6)

	t.Log("save event for another aggregate")
	id2, _ := eh.ParseUUID("c1138e5e-f6fb-4dd0-8e79-255c6c8d3756")
	agg2 := mocks.NewAggregate(id2)
	event7 := agg2.NewEvent(mocks.EventType, &mocks.EventData{"event7"})
	agg2.ApplyEvent(ctx, event7) // Apply event to increment the aggregate version.
	err = store.Save(ctx, []eh.Event{event7}, 0)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	savedEvents = append(savedEvents, event7)

	t.Log("load events for non-existing aggregate")
	events, err := store.Load(ctx, mocks.AggregateType, eh.NewUUID())
	if err != nil {
		t.Error("there should be no error:", err)
	}
	if len(events) != 0 {
		t.Error("there should be no loaded events:", eventsToString(events))
	}

	t.Log("load events")
	events, err = store.Load(ctx, mocks.AggregateType, id)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	expectedEvents := []eh.Event{
		event1,                 // Version 1
		event2,                 // Version 2
		event3,                 // Version 3
		event4, event5, event6, // Version 4, 5 and 6
	}
	for i, event := range events {
		if err := mocks.CompareEvents(event, expectedEvents[i]); err != nil {
			t.Error("the event was incorrect:", err)
		}
		if event.Version() != i+1 {
			t.Error("the event version should be correct:", event, event.Version())
		}
	}
	t.Log(events)

	t.Log("load events for another aggregate")
	events, err = store.Load(ctx, mocks.AggregateType, id2)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	expectedEvents = []eh.Event{event7}
	for i, event := range events {
		if err := mocks.CompareEvents(event, expectedEvents[i]); err != nil {
			t.Error("the event was incorrect:", err)
		}
		if event.Version() != i+1 {
			t.Error("the event version should be correct:", event, event.Version())
		}
	}

	return savedEvents
}

func eventsToString(events []eh.Event) string {
	parts := make([]string, len(events))
	for i, e := range events {
		parts[i] = string(e.AggregateType()) +
			":" + string(e.EventType()) +
			" (" + e.AggregateID().String() + ")"
	}
	return strings.Join(parts, ", ")
}
