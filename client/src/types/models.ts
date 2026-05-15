export type UserRole = 'user' | 'moderator';

export type User = {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  city: string;
  about: string;
  faculty: string;
  course: string;
  avatarUri: string;
};

export type EventItem = {
  id: string;
  title: string;
  shortDescription: string;
  description: string;
  date: string;
  location: string;
  tags: string[];
  imageUri: string;
  attendees: string[];
  creatorId: string;
};

export type AuthPayload = {
  email: string;
  password: string;
};

export type RegisterPayload = AuthPayload & {
  name: string;
  city: string;
  role: UserRole;
};

export type CreateEventPayload = {
  title: string;
  description: string;
  date: string;
  location: string;
  tags: string[];
  imageUri: string;
};
