import { PayloadAction } from '@reduxjs/toolkit';
import { call, debounce, put, select, takeLatest } from 'redux-saga/effects';
import { createEventApi, fetchEventsApi, registerToEventApi, unregisterFromEventApi } from '../../api/api';
import { CreateEventPayload, EventItem } from '../../types/models';
import type { RootState } from '../../redux/store';
import {
  createEventRequest,
  eventsFailure,
  fetchEventsFailure,
  fetchEventsRequest,
  fetchEventsSuccess,
  registerForEventRequest,
} from '../slices/eventsSlice';

function* fetchEventsWorker() {
  try {
    const searchText: string = yield select((state: RootState) => state.events.searchText);
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi, searchText);
    yield put(fetchEventsSuccess(events));
  } catch (error) {
    yield put(fetchEventsFailure(error instanceof Error ? error.message : 'Ошибка загрузки мероприятий'));
  }
}

function* toggleRegistrationWorker(action: PayloadAction<string>) {
  try {
    const userId: string | undefined = yield select((state: RootState) => state.auth.user?.id);
    const currentEvents: EventItem[] = yield select((state: RootState) => state.events.list);
    const eventItem = currentEvents.find((item) => item.id === action.payload);
    if (!userId) {
      throw new Error('Требуется авторизация');
    }
    if (!eventItem) {
      throw new Error('Мероприятие не найдено');
    }
    if (eventItem.attendees.includes(userId)) {
      yield call(unregisterFromEventApi, action.payload);
    } else {
      yield call(registerToEventApi, action.payload);
    }
    const searchText: string = yield select((state: RootState) => state.events.searchText);
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi, searchText);
    yield put(fetchEventsSuccess(events));
  } catch (error) {
    yield put(eventsFailure(error instanceof Error ? error.message : 'Ошибка регистрации на мероприятие'));
  }
}

function* createEventWorker(action: PayloadAction<CreateEventPayload>) {
  try {
    const userId: string | undefined = yield select((state: RootState) => state.auth.user?.id);
    if (!userId) {
      throw new Error('Нельзя создать мероприятие без авторизации');
    }
    yield call(createEventApi, action.payload);
    const searchText: string = yield select((state: RootState) => state.events.searchText);
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi, searchText);
    yield put(fetchEventsSuccess(events));
  } catch (error) {
    yield put(eventsFailure(error instanceof Error ? error.message : 'Ошибка создания мероприятия'));
  }
}

export function* watchEventsSaga() {
  yield debounce(400, fetchEventsRequest.type, fetchEventsWorker);
  yield takeLatest(registerForEventRequest.type, toggleRegistrationWorker);
  yield takeLatest(createEventRequest.type, createEventWorker);
}
