import { PayloadAction } from '@reduxjs/toolkit';
import { call, put, select, takeLatest } from 'redux-saga/effects';
import { createEventApi, fetchEventsApi, registerToEventApi } from '../../api/api';
import { CreateEventPayload } from '../../types/models';
import type { RootState } from '../../redux/store';
import {
  createEventRequest,
  fetchEventsFailure,
  fetchEventsRequest,
  fetchEventsSuccess,
  registerForEventRequest,
  updateEvents,
} from '../slices/eventsSlice';

function* fetchEventsWorker() {
  try {
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi);
    yield put(fetchEventsSuccess(events));
  } catch (error) {
    yield put(fetchEventsFailure(error instanceof Error ? error.message : 'Ошибка загрузки мероприятий'));
  }
}

function* toggleRegistrationWorker(action: PayloadAction<string>) {
  try {
    const userId: string | undefined = yield select((state: RootState) => state.auth.user?.id);
    if (!userId) {
      throw new Error('Требуется авторизация');
    }
    const events: Awaited<ReturnType<typeof registerToEventApi>> = yield call(registerToEventApi, action.payload, userId);
    yield put(updateEvents(events));
  } catch (error) {
    yield put(fetchEventsFailure(error instanceof Error ? error.message : 'Ошибка регистрации на мероприятие'));
  }
}

function* createEventWorker(action: PayloadAction<CreateEventPayload>) {
  try {
    const userId: string | undefined = yield select((state: RootState) => state.auth.user?.id);
    if (!userId) {
      throw new Error('Нельзя создать мероприятие без авторизации');
    }
    const events: Awaited<ReturnType<typeof createEventApi>> = yield call(createEventApi, action.payload, userId);
    yield put(updateEvents(events));
  } catch (error) {
    yield put(fetchEventsFailure(error instanceof Error ? error.message : 'Ошибка создания мероприятия'));
  }
}

export function* watchEventsSaga() {
  yield takeLatest(fetchEventsRequest.type, fetchEventsWorker);
  yield takeLatest(registerForEventRequest.type, toggleRegistrationWorker);
  yield takeLatest(createEventRequest.type, createEventWorker);
}
