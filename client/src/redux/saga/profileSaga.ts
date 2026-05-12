import { PayloadAction } from '@reduxjs/toolkit';
import { call, put, select, takeLatest } from 'redux-saga/effects';
import { updateProfileApi } from '../../api/api';
import { User } from '../../types/models';
import { updateAuthUser } from '../slices/authSlice';
import {
  profileFailure,
  updateAvatarRequest,
  updateProfileRequest,
  updateProfileSuccess,
} from '../slices/profileSlice';
import type { RootState } from '../../redux/store';

function* updateProfileWorker(action: PayloadAction<Partial<User>>) {
  try {
    const userId: string | undefined = yield select((state: RootState) => state.auth.user?.id);
    if (!userId) {
      throw new Error('Профиль недоступен');
    }
    const profile: Awaited<ReturnType<typeof updateProfileApi>> = yield call(updateProfileApi, userId, action.payload);
    yield put(updateProfileSuccess(profile));
    yield put(updateAuthUser(profile));
  } catch (error) {
    yield put(profileFailure(error instanceof Error ? error.message : 'Ошибка обновления профиля'));
  }
}

function* updateAvatarWorker(action: PayloadAction<string>) {
  try {
    const userId: string | undefined = yield select((state: RootState) => state.auth.user?.id);
    if (!userId) {
      throw new Error('Профиль недоступен');
    }
    const profile: Awaited<ReturnType<typeof updateProfileApi>> = yield call(updateProfileApi, userId, {
      avatarUri: action.payload,
    });
    yield put(updateProfileSuccess(profile));
    yield put(updateAuthUser(profile));
  } catch (error) {
    yield put(profileFailure(error instanceof Error ? error.message : 'Ошибка обновления аватара'));
  }
}

export function* watchProfileSaga() {
  yield takeLatest(updateProfileRequest.type, updateProfileWorker);
  yield takeLatest(updateAvatarRequest.type, updateAvatarWorker);
}
