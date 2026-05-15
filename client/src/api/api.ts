import axios, { AxiosError } from 'axios';
import Constants from 'expo-constants';
import { Platform } from 'react-native';
import { AuthPayload, CreateEventPayload, EventItem, RegisterPayload, User } from '../types/models';

type AuthResponse = {
  access_token: string;
  refresh_token: string;
  expires_at: string;
};

type ServerUserResponse = {
  id: number;
  name: string;
  email: string;
  role: 'user' | 'moderator';
  about?: string | null;
  city: string;
  faculty?: string | null;
  course?: string | null;
  avatar_image_id?: string | null;
};

type ServerAvatarResponse = {
  data_url?: string;
  url?: string;
};

type ServerEventResponse = {
  id: number;
  title: string;
  description: string;
  event_date: string;
  location: string;
  image_id?: string | null;
  creator_id: number;
  attendees: number[];
  tags: string[];
  created_at: string;
  updated_at: string;
};

type ServerEventImageResponse = {
  image_id: string;
  url: string;
};

type ServerEventsResponse = {
  events: ServerEventResponse[];
  count: number;
  limit: number;
  offset: number;
};

const extractHost = (rawValue?: string | null) => {
  const value = rawValue?.trim();
  if (!value) {
    return '';
  }

  const withoutProtocol = value.replace(/^(exp|exps|http|https):\/\//, '');
  const [hostPart] = withoutProtocol.split('/');
  const [host] = hostPart.split(':');
  return host?.trim() ?? '';
};

const resolveApiBaseUrl = () => {
  const envUrl = process.env.EXPO_PUBLIC_API_URL?.trim();
  if (envUrl) {
    return envUrl.replace(/\/$/, '');
  }

  const hostCandidates = [
    Constants.expoConfig?.hostUri,
    (Constants as { manifest?: { debuggerHost?: string } }).manifest?.debuggerHost,
    (Constants as { expoGoConfig?: { debuggerHost?: string } }).expoGoConfig?.debuggerHost,
    (Constants as { manifest2?: { extra?: { expoClient?: { hostUri?: string } } } }).manifest2?.extra?.expoClient?.hostUri,
  ];

  for (const candidate of hostCandidates) {
    const host = extractHost(candidate);
    if (host) {
      return `http://${host}:8080`;
    }
  }

  if (Platform.OS === 'android') {
    return 'http://10.0.2.2:8080';
  }

  return 'http://localhost:8080';
};

const apiBaseUrl = resolveApiBaseUrl();
console.log('[api] resolved base url', apiBaseUrl);

const apiClient = axios.create({
  baseURL: apiBaseUrl,
  timeout: 10000,
});

let accessToken: string | null = null;

export const setApiAccessToken = (token: string | null) => {
  accessToken = token;
};

apiClient.interceptors.request.use((config) => {
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  console.log('[api] request', {
    method: config.method,
    url: `${config.baseURL ?? ''}${config.url ?? ''}`,
  });
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error)) {
      console.log('[api] response error', {
        code: error.code,
        message: error.message,
        status: error.response?.status ?? null,
        url: `${error.config?.baseURL ?? ''}${error.config?.url ?? ''}`,
      });
    }
    return Promise.reject(error);
  }
);

const toApiError = (error: unknown, fallbackMessage: string) => {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<{ error?: string; message?: string }>;
    const responseError = axiosError.response?.data?.error ?? axiosError.response?.data?.message;
    return new Error(responseError || axiosError.message || fallbackMessage);
  }
  if (error instanceof Error) {
    return new Error(error.message);
  }
  return new Error(fallbackMessage);
};

const serverUserToClient = (user: ServerUserResponse, avatarUri = ''): User => ({
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

const serverEventToClient = (eventItem: ServerEventResponse): EventItem => ({
  id: String(eventItem.id),
  title: eventItem.title,
  description: eventItem.description,
  date: eventItem.event_date.slice(0, 10),
  location: eventItem.location,
  tags: eventItem.tags ?? [],
  imageUri: buildEventImageUri(eventItem),
  attendees: (eventItem.attendees ?? []).map(String),
  creatorId: String(eventItem.creator_id),
  shortDescription: buildShortDescription(eventItem.description),
});

const buildShortDescription = (description: string) => {
  const text = description.trim();
  if (!text) {
    return '';
  }

  const maxLength = 140;
  if (text.length <= maxLength) {
    return text;
  }

  return `${text.slice(0, maxLength).trimEnd()}...`;
};

const buildEventImageUri = (eventItem: ServerEventResponse) => {
  const imageId = eventItem.image_id?.trim();
  if (!imageId) {
    return '';
  }
  if (imageId.startsWith('http://') || imageId.startsWith('https://') || imageId.startsWith('data:')) {
    return imageId;
  }
  return new URL(`/api/v1/events/images/${imageId}/file`, apiBaseUrl).toString();
};

const buildEventPayload = (payload: CreateEventPayload) => ({
  title: payload.title,
  description: payload.description,
  event_date: payload.date,
  location: payload.location,
  tags: payload.tags,
  image_uri: payload.imageUri,
});

export const loginApi = async (payload: AuthPayload): Promise<AuthResponse> => {
  try {
    const response = await apiClient.post<AuthResponse>('/api/v1/auth/login', payload);
    return response.data;
  } catch (error) {
    throw toApiError(error, 'Ошибка входа');
  }
};

export const registerApi = async (payload: RegisterPayload): Promise<AuthResponse> => {
  try {
    const response = await apiClient.post<AuthResponse>('/api/v1/auth/register', payload);
    return response.data;
  } catch (error) {
    throw toApiError(error, 'Ошибка регистрации');
  }
};

export const fetchCurrentUserApi = async (): Promise<ServerUserResponse> => {
  try {
    const response = await apiClient.get<ServerUserResponse>('/api/v1/users/me');
    return response.data;
  } catch (error) {
    throw toApiError(error, 'Не удалось загрузить профиль');
  }
};

export const fetchUserAvatarApi = async (userId: string): Promise<string> => {
  try {
    const response = await apiClient.get<ServerAvatarResponse>(`/api/v1/users/${userId}/avatar`);
    console.log('[avatar] fetch response', {
      userId,
      status: response.status,
      hasDataUrl: Boolean(response.data.data_url),
      dataUrlLength: response.data.data_url?.length ?? 0,
      hasUrl: Boolean(response.data.url),
      url: response.data.url ?? '',
    });

    const dataUrl = response.data.data_url?.trim();
    if (dataUrl) {
      console.log('[avatar] using data_url', {
        userId,
        prefix: dataUrl.slice(0, 32),
        length: dataUrl.length,
      });
      return dataUrl;
    }

    const publicFileUrl = new URL(`/api/v1/users/${userId}/avatar/file`, apiBaseUrl).toString();
    console.log('[avatar] using public file fallback', {
      userId,
      prefix: publicFileUrl.slice(0, 64),
      length: publicFileUrl.length,
    });
    return publicFileUrl;
  } catch (error) {
    throw toApiError(error, 'Не удалось загрузить аватар');
  }
};

export const updateProfileApi = async (payload: Partial<User>): Promise<User> => {
  try {
    const response = await apiClient.put<ServerUserResponse>('/api/v1/users/me', {
      name: payload.name,
      about: payload.about,
      city: payload.city,
      faculty: payload.faculty,
      course: payload.course,
    });
    let avatarUri = '';
    try {
      avatarUri = await fetchUserAvatarApi(String(response.data.id));
    } catch {
      avatarUri = '';
    }
    return serverUserToClient(response.data, avatarUri);
  } catch (error) {
    throw toApiError(error, 'Не удалось обновить профиль');
  }
};

export const updateAvatarApi = async (uri: string): Promise<User> => {
  try {
    console.log('[avatar] upload start', {
      uriPrefix: uri.slice(0, 64),
      uriLength: uri.length,
    });
    const fileName = uri.split('/').pop() || 'avatar.jpg';
    const match = /\.([a-zA-Z0-9]+)$/.exec(fileName);
    const type = match ? `image/${match[1].toLowerCase()}` : 'image/jpeg';
    const formData = new FormData();
    formData.append('avatar', {
      uri,
      name: fileName,
      type,
    } as never);

    const response = await fetch(`${resolveApiBaseUrl()}/api/v1/users/me/avatar`, {
      method: 'PUT',
      headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
      body: formData,
    });

    if (!response.ok) {
      const errorBody = (await response.json().catch(() => ({}))) as { error?: string; message?: string };
      console.log('[avatar] upload failed', {
        status: response.status,
        error: errorBody.error ?? errorBody.message ?? '',
      });
      throw new Error(errorBody.error || errorBody.message || 'Не удалось обновить аватар');
    }

    const responseData = (await response.json()) as ServerUserResponse;
    console.log('[avatar] upload response user', {
      id: responseData.id,
      avatar_image_id: responseData.avatar_image_id ?? null,
    });
    let avatarUri = '';
    try {
      avatarUri = await fetchUserAvatarApi(String(responseData.id));
    } catch {
      avatarUri = '';
    }
    console.log('[avatar] upload resolved uri', {
      id: responseData.id,
      hasAvatarUri: Boolean(avatarUri),
      prefix: avatarUri.slice(0, 64),
      length: avatarUri.length,
    });
    return serverUserToClient(responseData, avatarUri);
  } catch (error) {
    throw toApiError(error, 'Не удалось обновить аватар');
  }
};

export const deleteAvatarApi = async (): Promise<User> => {
  try {
    const response = await apiClient.delete<ServerUserResponse>('/api/v1/users/me/avatar');
    return serverUserToClient(response.data, '');
  } catch (error) {
    throw toApiError(error, 'Не удалось удалить аватар');
  }
};

export const fetchEventsApi = async (search = ''): Promise<EventItem[]> => {
  try {
    const params = search.trim() ? { search: search.trim() } : undefined;
    const response = await apiClient.get<ServerEventsResponse>('/api/v1/events', { params });
    return response.data.events.map(serverEventToClient);
  } catch (error) {
    throw toApiError(error, 'Не удалось загрузить мероприятия');
  }
};

export const registerToEventApi = async (eventId: string): Promise<void> => {
  try {
    await apiClient.post(`/api/v1/events/${eventId}/register`);
  } catch (error) {
    throw toApiError(error, 'Не удалось записаться на мероприятие');
  }
};

export const unregisterFromEventApi = async (eventId: string): Promise<void> => {
  try {
    await apiClient.delete(`/api/v1/events/${eventId}/register`);
  } catch (error) {
    throw toApiError(error, 'Не удалось отменить регистрацию');
  }
};

export const createEventApi = async (payload: CreateEventPayload): Promise<void> => {
  try {
    await apiClient.post('/api/v1/events', buildEventPayload(payload));
  } catch (error) {
    throw toApiError(error, 'Не удалось создать мероприятие');
  }
};

export const uploadEventImageApi = async (uri: string): Promise<ServerEventImageResponse> => {
  try {
    const fileName = uri.split('/').pop() || 'event-image.jpg';
    const match = /\.([a-zA-Z0-9]+)$/.exec(fileName);
    const type = match ? `image/${match[1].toLowerCase()}` : 'image/jpeg';
    const formData = new FormData();
    formData.append('image', {
      uri,
      name: fileName,
      type,
    } as never);

    const response = await fetch(`${apiBaseUrl}/api/v1/events/image`, {
      method: 'POST',
      headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
      body: formData,
    });

    if (!response.ok) {
      const errorBody = (await response.json().catch(() => ({}))) as { error?: string; message?: string };
      throw new Error(errorBody.error || errorBody.message || 'Не удалось загрузить обложку мероприятия');
    }

    return (await response.json()) as ServerEventImageResponse;
  } catch (error) {
    throw toApiError(error, 'Не удалось загрузить обложку мероприятия');
  }
};
