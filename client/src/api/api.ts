import axios from 'axios';
import { AuthPayload, CreateEventPayload, EventItem, RegisterPayload, User } from '../types/models';

const apiClient = axios.create({
  baseURL: 'https://student-eventor.local',
  timeout: 3000,
});

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

const mockUsers: User[] = [
  {
    id: 'u1',
    name: 'Ирина Волкова',
    email: 'user@eventor.ru',
    role: 'user',
    about: 'Студентка 3 курса, люблю хакатоны и дизайн.',
    faculty: 'ИВТ',
    course: '3',
    avatarUri: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=400',
  },
  {
    id: 'm1',
    name: 'Артем Лебедев',
    email: 'moderator@eventor.ru',
    role: 'moderator',
    about: 'Модератор студенческого клуба, организую мероприятия.',
    faculty: 'ИС',
    course: '4',
    avatarUri: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=400',
  },
];

let eventsDb: EventItem[] = [
  {
    id: 'e1',
    title: 'Ночь проектирования',
    shortDescription: 'Командная работа над учебными pet-проектами.',
    description: 'Открытая встреча для студентов: команды, трекеры, мини-защита в конце.',
    date: '2026-05-25',
    location: 'Коворкинг А-311',
    tags: ['IT', 'Проекты', 'Команды'],
    imageUri: 'https://images.unsplash.com/photo-1521737604893-d14cc237f11d?w=800',
    attendees: ['u1'],
    creatorId: 'm1',
  },
  {
    id: 'e2',
    title: 'Ораторский интенсив',
    shortDescription: 'Практика публичных выступлений для студентов.',
    description: 'Разберем структуру доклада, подачу и работу со страхом сцены.',
    date: '2026-06-02',
    location: 'Актовый зал',
    tags: ['Soft Skills', 'Выступления'],
    imageUri: 'https://images.unsplash.com/photo-1515169067868-5387ec356754?w=800',
    attendees: ['u1'],
    creatorId: 'm1',
  },
  {
    id: 'e3',
    title: 'GameDev Jam',
    shortDescription: '48 часов на создание прототипа игры.',
    description: 'Интенсив с менторами, художниками и разработчиками.',
    date: '2026-06-15',
    location: 'Технопарк, аудитория 7',
    tags: ['GameDev', 'Hackathon'],
    imageUri: 'https://images.unsplash.com/photo-1542751371-adc38448a05e?w=800',
    attendees: [],
    creatorId: 'm1',
  },
];

const wrap = async <T,>(data: T): Promise<T> => {
  await sleep(350);
  await apiClient.get('/health', {
    adapter: async (config) => ({
      data: { ok: true },
      status: 200,
      statusText: 'OK',
      headers: {},
      config,
    }),
  });
  return data;
};

export const loginApi = async (payload: AuthPayload): Promise<{ token: string; user: User }> => {
  const user = mockUsers.find((item) => item.email.toLowerCase() === payload.email.trim().toLowerCase());
  if (!user) {
    throw new Error('Пользователь не найден. Используйте user@eventor.ru или moderator@eventor.ru');
  }
  return wrap({ token: `token-${user.id}`, user });
};

export const registerApi = async (payload: RegisterPayload): Promise<{ token: string; user: User }> => {
  const existing = mockUsers.find((item) => item.email.toLowerCase() === payload.email.trim().toLowerCase());
  if (existing) {
    throw new Error('Пользователь с такой почтой уже существует');
  }
  const newUser: User = {
    id: `u${mockUsers.length + 1}`,
    name: payload.name.trim(),
    email: payload.email.trim(),
    role: payload.role,
    about: 'Расскажите о себе',
    faculty: 'Не указано',
    course: '1',
    avatarUri: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=400',
  };
  mockUsers.push(newUser);
  return wrap({ token: `token-${newUser.id}`, user: newUser });
};

export const updateProfileApi = async (userId: string, payload: Partial<User>): Promise<User> => {
  const user = mockUsers.find((item) => item.id === userId);
  if (!user) {
    throw new Error('Профиль не найден');
  }
  Object.assign(user, payload);
  return wrap({ ...user });
};

export const fetchEventsApi = async (): Promise<EventItem[]> => {
  return wrap([...eventsDb]);
};

export const registerToEventApi = async (eventId: string, userId: string): Promise<EventItem[]> => {
  eventsDb = eventsDb.map((eventItem) => {
    if (eventItem.id !== eventId) {
      return eventItem;
    }
    const hasRegistration = eventItem.attendees.includes(userId);
    return {
      ...eventItem,
      attendees: hasRegistration
        ? eventItem.attendees.filter((attendeeId) => attendeeId !== userId)
        : [...eventItem.attendees, userId],
    };
  });
  return wrap([...eventsDb]);
};

export const createEventApi = async (payload: CreateEventPayload, creatorId: string): Promise<EventItem[]> => {
  const newEvent: EventItem = {
    id: `e${eventsDb.length + 1}`,
    creatorId,
    attendees: [],
    ...payload,
  };
  eventsDb = [newEvent, ...eventsDb];
  return wrap([...eventsDb]);
};
