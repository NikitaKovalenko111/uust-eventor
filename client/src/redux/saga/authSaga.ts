import { PayloadAction } from '@reduxjs/toolkit';
import { call, put, takeLatest } from 'redux-saga/effects';
import { loginApi, registerApi } from '../../api/api';
import { AuthPayload, RegisterPayload } from '../../types/models';
import { authFailure, authSuccess, loginRequest, registerRequest } from '../slices/authSlice';
import { setProfile } from '../slices/profileSlice';

function* loginWorker(action: PayloadAction<AuthPayload>) {
  try {
    const data: { token: string; user: Awaited<ReturnType<typeof loginApi>>['user'] } = yield call(loginApi, action.payload);
    yield put(authSuccess(data));
    yield put(setProfile(data.user));
  } catch (error) {
    yield put(authFailure(error instanceof Error ? error.message : 'Ошибка входа'));
  }
}

function* registerWorker(action: PayloadAction<RegisterPayload>) {
  try {
    const data: { token: string; user: Awaited<ReturnType<typeof registerApi>>['user'] } = yield call(registerApi, action.payload);
    yield put(authSuccess(data));
    yield put(setProfile(data.user));
  } catch (error) {
    yield put(authFailure(error instanceof Error ? error.message : 'Ошибка регистрации'));
  }
}

export function* watchAuthSaga() {
  yield takeLatest(loginRequest.type, loginWorker);
  yield takeLatest(registerRequest.type, registerWorker);
}
