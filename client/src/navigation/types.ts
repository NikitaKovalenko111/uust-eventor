export type RootStackParamList = {
  Auth: undefined;
  UserCabinet: undefined;
  ModeratorCabinet: undefined;
  Events: undefined;
  UserSearch: undefined;
  FriendRequests: undefined;
  EventDetails: { eventId: string };
  Profile: { userId: string };
};
