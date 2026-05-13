import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { User } from '../../types/models';

type ProfileState = {
  profile: User | null;
  loading: boolean;
  error: string | null;
};

const initialState: ProfileState = {
  profile: null,
  loading: false,
  error: null,
};

const profileSlice = createSlice({
  name: 'profile',
  initialState,
  reducers: {
    setProfile(state, action: PayloadAction<User | null>) {
      state.profile = action.payload;
    },
    updateProfileRequest(state, _action: PayloadAction<Partial<User>>) {
      state.loading = true;
      state.error = null;
    },
    updateAvatarRequest(state, action: PayloadAction<string>) {
      state.loading = true;
      state.error = null;
      if (state.profile) {
        state.profile.avatarUri = action.payload;
      }
    },
    updateProfileSuccess(state, action: PayloadAction<User>) {
      state.loading = false;
      state.profile = action.payload;
    },
    profileFailure(state, action: PayloadAction<string>) {
      state.loading = false;
      state.error = action.payload;
    },
  },
});

export const {
  setProfile,
  updateProfileRequest,
  updateAvatarRequest,
  updateProfileSuccess,
  profileFailure,
} = profileSlice.actions;

export default profileSlice.reducer;
