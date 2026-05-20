import { all, fork } from 'redux-saga/effects';
import { watchAuthSaga } from './authSaga';
import { watchEventsSaga } from './eventsSaga';
import { watchProfileSaga } from './profileSaga';

export function* rootSaga() {
  yield all([fork(watchAuthSaga), fork(watchProfileSaga), fork(watchEventsSaga)]);
}
