import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { AuthPayload, RegisterPayload, User } from '../../types/models';

type AuthState = {
  isAuthenticated: boolean;
  token: string | null;
  refreshToken: string | null;
  user: User | null;
  loading: boolean;
  error: string | null;
};

const initialState: AuthState = {
  isAuthenticated: false,
  token: null,
  refreshToken: null,
  user: null,
  loading: false,
  error: null,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    loginRequest(state, _action: PayloadAction<AuthPayload>) {
      state.loading = true;
      state.error = null;
    },
    registerRequest(state, _action: PayloadAction<RegisterPayload>) {
      state.loading = true;
      state.error = null;
    },
    authSuccess(state, action: PayloadAction<{ token: string; refreshToken: string; user: User }>) {
      state.loading = false;
      state.isAuthenticated = true;
      state.token = action.payload.token;
      state.refreshToken = action.payload.refreshToken;
      state.user = action.payload.user;
      state.error = null;
    },
    authFailure(state, action: PayloadAction<string>) {
      state.loading = false;
      state.error = action.payload;
    },
    updateAuthUser(state, action: PayloadAction<User>) {
      state.user = action.payload;
    },
    logout(state) {
      state.isAuthenticated = false;
      state.token = null;
      state.refreshToken = null;
      state.user = null;
      state.loading = false;
      state.error = null;
    },
  },
});

export const {
  loginRequest,
  registerRequest,
  authSuccess,
  authFailure,
  updateAuthUser,
  logout,
} = authSlice.actions;

export default authSlice.reducer;
