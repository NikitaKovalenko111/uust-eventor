import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { CreateEventPayload, EventItem } from '../../types/models';

type EventsState = {
  list: EventItem[];
  loading: boolean;
  error: string | null;
  searchText: string;
};

const initialState: EventsState = {
  list: [],
  loading: false,
  error: null,
  searchText: '',
};

const eventsSlice = createSlice({
  name: 'events',
  initialState,
  reducers: {
    fetchEventsRequest(state) {
      console.log("state");
      
      state.loading = true;
      state.error = null;
    },
    fetchEventsSuccess(state, action: PayloadAction<EventItem[]>) {
      state.loading = false;
      state.list = action.payload;
    },
    fetchEventsFailure(state, action: PayloadAction<string>) {
      state.loading = false;
      state.error = action.payload;
    },
    eventsFailure(state, action: PayloadAction<string>) {
      state.loading = false;
      state.error = action.payload;
    },
    registerForEventRequest(state, _action: PayloadAction<string>) {
      state.loading = true;
      state.error = null;
    },
    createEventRequest(state, _action: PayloadAction<CreateEventPayload>) {
      state.loading = true;
      state.error = null;
    },
    updateEvents(state, action: PayloadAction<EventItem[]>) {
      state.loading = false;
      state.list = action.payload;
    },
    setSearchText(state, action: PayloadAction<string>) {
      state.searchText = action.payload;
    },
    clearFilters(state) {
      state.searchText = '';
    },
  },
});

export const {
  fetchEventsRequest,
  fetchEventsSuccess,
  fetchEventsFailure,
  eventsFailure,
  registerForEventRequest,
  createEventRequest,
  updateEvents,
  setSearchText,
  clearFilters,
} = eventsSlice.actions;

export default eventsSlice.reducer;
