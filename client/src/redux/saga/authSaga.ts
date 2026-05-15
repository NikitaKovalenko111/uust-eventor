import { PayloadAction } from '@reduxjs/toolkit';
import { call, put, takeLatest } from 'redux-saga/effects';
import {
  fetchCurrentUserApi,
  fetchEventsApi,
  fetchUserAvatarApi,
  loginApi,
  registerApi,
  setApiAccessToken,
} from '../../api/api';
import { AuthPayload, RegisterPayload } from '../../types/models';
import { authFailure, authSuccess, loginRequest, registerRequest } from '../slices/authSlice';
import { setProfile } from '../slices/profileSlice';
import { fetchEventsSuccess } from '../slices/eventsSlice';
import { User } from '../../types/models';

const toClientUser = (user: Awaited<ReturnType<typeof fetchCurrentUserApi>>, avatarUri: string): User => ({
  id: String(user.id),
  name: user.name,
  email: user.email,
  role: user.role,
  city: user.city,
  about: user.about ?? '',
  faculty: user.faculty ?? '',
  course: user.course ?? '',
  avatarUri,
});

function* loginWorker(action: PayloadAction<AuthPayload>): Generator {
  try {
    const data: Awaited<ReturnType<typeof loginApi>> = yield call(loginApi, action.payload);
    yield call(setApiAccessToken, data.access_token);

    const userResponse: Awaited<ReturnType<typeof fetchCurrentUserApi>> = yield call(fetchCurrentUserApi);
    let avatarUri = '';
    try {
      avatarUri = yield call(fetchUserAvatarApi, String(userResponse.id));
    } catch {
      avatarUri = '';
    }

    const user = toClientUser(userResponse, avatarUri);
    yield put(authSuccess({ token: data.access_token, refreshToken: data.refresh_token, user }));
    yield put(setProfile(user));
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi);
    yield put(fetchEventsSuccess(events));
  } catch (error) {
    yield put(authFailure(error instanceof Error ? error.message : 'Ошибка входа'));
  }
}

function* registerWorker(action: PayloadAction<RegisterPayload>): Generator {
  try {
    const data: Awaited<ReturnType<typeof registerApi>> = yield call(registerApi, action.payload);
    yield call(setApiAccessToken, data.access_token);
    const userResponse: Awaited<ReturnType<typeof fetchCurrentUserApi>> = yield call(fetchCurrentUserApi);
    let avatarUri = '';
    try {
      avatarUri = yield call(fetchUserAvatarApi, String(userResponse.id));
    } catch {
      avatarUri = '';
    }
    const user = toClientUser(userResponse, avatarUri);
    yield put(authSuccess({ token: data.access_token, refreshToken: data.refresh_token, user }));
    yield put(setProfile(user));
    const events: Awaited<ReturnType<typeof fetchEventsApi>> = yield call(fetchEventsApi);
    yield put(fetchEventsSuccess(events));
  } catch (error) {
    yield put(authFailure(error instanceof Error ? error.message : 'Ошибка регистрации'));
  }
}

export function* watchAuthSaga() {
  yield takeLatest(loginRequest.type, loginWorker);
  yield takeLatest(registerRequest.type, registerWorker);
}
