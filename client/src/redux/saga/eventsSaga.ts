import { PayloadAction } from '@reduxjs/toolkit';
import { call, debounce, put, select, takeLatest } from 'redux-saga/effects';
import { createEventApi, fetchEventsApi, registerToEventApi, unregisterFromEventApi, deleteEventApi, finishEventApi } from '../../api/api';
import { CreateEventPayload, EventItem } from '../../types/models';
import type { RootState } from '../../redux/store';
import {
  createEventRequest,
  eventsFailure,
  fetchEventsFailure,
  fetchEventsRequest,
  fetchEventsSuccess,
  registerForEventRequest,
  deleteEventRequest,
  finishEventRequest,
} from '../slices/eventsSlice';

function* fetchEventsWorker() {
  try {
    const searchText: string = yield select((state: RootState) => state.events.searchText);
    const limit: number = yield select((state: RootState) => state.events.limit);
    const offset: number = yield select((state: RootState) => state.events.offset);
    const city: string = yield select((state: RootState) => state.auth.user?.city ?? state.profile.profile?.city ?? '');
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi, searchText, limit, offset, city);
    yield put(fetchEventsSuccess(sortEventsByCityRelevance(events, city)));
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
    const limit: number = yield select((state: RootState) => state.events.limit);
    const offset: number = yield select((state: RootState) => state.events.offset);
    const city: string = yield select((state: RootState) => state.auth.user?.city ?? state.profile.profile?.city ?? '');
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi, searchText, limit, offset, city);
    yield put(fetchEventsSuccess(sortEventsByCityRelevance(events, city)));
  } catch (error) {
    yield put(eventsFailure(error instanceof Error ? error.message : 'Ошибка регистрации на мероприятие'));
  }
}

function* deleteEventWorker(action: PayloadAction<string>) {
  try {
    yield call(deleteEventApi, action.payload);
    const searchText: string = yield select((state: RootState) => state.events.searchText);
    const limit: number = yield select((state: RootState) => state.events.limit);
    const offset: number = yield select((state: RootState) => state.events.offset);
    const city: string = yield select((state: RootState) => state.auth.user?.city ?? state.profile.profile?.city ?? '');
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi, searchText, limit, offset, city);
    yield put(fetchEventsSuccess(sortEventsByCityRelevance(events, city)));
  } catch (error) {
    yield put(eventsFailure(error instanceof Error ? error.message : 'Ошибка удаления мероприятия'));
  }
}

function* finishEventWorker(action: PayloadAction<string>) {
  try {
    yield call(finishEventApi, action.payload);
    const searchText: string = yield select((state: RootState) => state.events.searchText);
    const limit: number = yield select((state: RootState) => state.events.limit);
    const offset: number = yield select((state: RootState) => state.events.offset);
    const city: string = yield select((state: RootState) => state.auth.user?.city ?? state.profile.profile?.city ?? '');
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi, searchText, limit, offset, city);
    yield put(fetchEventsSuccess(sortEventsByCityRelevance(events, city)));
  } catch (error) {
    yield put(eventsFailure(error instanceof Error ? error.message : 'Ошибка завершения мероприятия'));
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
    const limit: number = yield select((state: RootState) => state.events.limit);
    const offset: number = yield select((state: RootState) => state.events.offset);
    const city: string = yield select((state: RootState) => state.auth.user?.city ?? state.profile.profile?.city ?? '');
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi, searchText, limit, offset, city);
    yield put(fetchEventsSuccess(sortEventsByCityRelevance(events, city)));
  } catch (error) {
    yield put(eventsFailure(error instanceof Error ? error.message : 'Ошибка создания мероприятия'));
  }
}

function sortEventsByCityRelevance(events: EventItem[], city: string) {
  const normalizedCity = city.trim().toLowerCase();
  if (!normalizedCity) {
    return [...events].sort((left, right) => (right.relevanceScore ?? 0) - (left.relevanceScore ?? 0));
  }

  return events
    .map((event, index) => ({
      event,
      index,
      cityRelevant: event.location.toLowerCase().includes(normalizedCity),
      score: event.relevanceScore ?? 0,
    }))
    .sort((left, right) => {
      if (left.cityRelevant !== right.cityRelevant) {
        return left.cityRelevant ? -1 : 1;
      }

      if (left.score !== right.score) {
        return right.score - left.score;
      }
      return left.index - right.index;
    })
    .map(({ event }) => event);
}

export function* watchEventsSaga() {
  yield debounce(400, fetchEventsRequest.type, fetchEventsWorker);
  yield takeLatest(registerForEventRequest.type, toggleRegistrationWorker);
  yield takeLatest(createEventRequest.type, createEventWorker);
  yield takeLatest(deleteEventRequest.type, deleteEventWorker);
  yield takeLatest(finishEventRequest.type, finishEventWorker);
}
