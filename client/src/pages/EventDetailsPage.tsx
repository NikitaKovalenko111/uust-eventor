import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { Image, StyleSheet, Text, View, Alert } from 'react-native';
import { AppHeader } from '../components/AppHeader';
import { PrimaryButton } from '../components/PrimaryButton';
import { ScreenContainer } from '../components/ScreenContainer';
import { theme } from '../constants/theme';
import { RootStackParamList } from '../navigation/types';
import { useAppDispatch, useAppSelector } from '../redux/hooks';
import { registerForEventRequest, deleteEventRequest, finishEventRequest } from '../redux/slices/eventsSlice';

type Props = NativeStackScreenProps<RootStackParamList, 'EventDetails'>;

export const EventDetailsPage = ({ route, navigation }: Props) => {
  const dispatch = useAppDispatch();
  const eventItem = useAppSelector((state) => state.events.list.find((event) => event.id === route.params.eventId));
  const userId = useAppSelector((state) => state.auth.user?.id ?? '');
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);

  if (!eventItem) {
    return (
      <ScreenContainer>
        <AppHeader title="Мероприятие" onActionPress={() => navigation.goBack()} actionIcon="arrow-back" />
        <Text style={styles.empty}>Мероприятие не найдено</Text>
      </ScreenContainer>
    );
  }

  const isRegistered = eventItem.attendees.includes(userId);
  const isOwnEvent = isAuthenticated && eventItem.creatorId === userId;

  return (
    <ScreenContainer>
      <AppHeader title="Страница мероприятия" onActionPress={() => navigation.goBack()} actionIcon="arrow-back" />

      {eventItem.imageUri ? (
        <View style={styles.photoBlock}>
          <Image source={{ uri: eventItem.imageUri }} style={styles.image} resizeMode="cover" />
        </View>
      ) : null}

      <Text style={styles.title}>{eventItem.title}</Text>
      <Text style={styles.shortDescription}>{eventItem.shortDescription}</Text>
      <Text style={styles.description}>{eventItem.description}</Text>
      <Text style={styles.meta}>Дата: {eventItem.date}</Text>
      <Text style={styles.meta}>Место: {eventItem.location}</Text>
      <Text style={styles.meta}>Теги: {eventItem.tags.join(' • ')}</Text>
      <Text style={styles.meta}>Участники: {eventItem.attendees.length}</Text>

      {!isOwnEvent ? (
        <PrimaryButton
          title={
            !isAuthenticated
              ? 'Войти, чтобы записаться'
              : isRegistered
                ? 'Отменить регистрацию'
                : 'Записаться на мероприятие'
          }
          type={!isAuthenticated || isRegistered ? 'outline' : 'primary'}
          onPress={() => {
            if (!isAuthenticated) {
              navigation.navigate('Auth');
              return;
            }
            dispatch(registerForEventRequest(eventItem.id));
          }}
        />
      ) : null}

      {isOwnEvent ? (
        <View style={{ marginTop: 12, gap: 8 }}>
          {!eventItem.finished ? (
            <PrimaryButton
              title="Завершить мероприятие"
              type="outline"
              onPress={() => {
                Alert.alert('Подтвердите', 'Вы уверены, что хотите завершить мероприятие?', [
                  { text: 'Отмена', style: 'cancel' },
                  { text: 'Да', onPress: () => dispatch(finishEventRequest(eventItem.id)) },
                ]);
              }}
            />
          ) : (
            <Text style={{ color: theme.colors.muted }}>Мероприятие завершено</Text>
          )}

          <PrimaryButton
            title="Удалить мероприятие"
            type="outline"
            onPress={() => {
              Alert.alert('Подтвердите', 'Вы действительно хотите удалить мероприятие?', [
                { text: 'Отмена', style: 'cancel' },
                {
                  text: 'Удалить',
                  style: 'destructive',
                  onPress: () => {
                    dispatch(deleteEventRequest(eventItem.id));
                    navigation.goBack();
                  },
                },
              ]);
            }}
          />
        </View>
      ) : null}

      <View style={styles.discussionBlock}>
        <Text style={styles.discussionTitle}>Блок обсуждения мероприятия</Text>
        <Text style={styles.discussionText}>Здесь можно добавить чат, комментарии и вопросы к организаторам.</Text>
      </View>
    </ScreenContainer>
  );
};

const styles = StyleSheet.create({
  photoBlock: {
    borderRadius: theme.radius.md,
    borderWidth: 1,
    borderColor: theme.colors.border,
    overflow: 'hidden',
    backgroundColor: theme.colors.surface,
  },
  image: {
    width: '100%',
    height: 190,
  },
  title: {
    marginTop: 4,
    color: theme.colors.text,
    fontSize: 20,
    fontWeight: '700',
  },
  shortDescription: {
    color: theme.colors.muted,
    fontSize: 14,
    fontWeight: '600',
  },
  description: {
    color: theme.colors.text,
    fontSize: 14,
    lineHeight: 22,
  },
  meta: {
    color: theme.colors.muted,
    fontSize: 13,
  },
  discussionBlock: {
    marginTop: theme.spacing.sm,
    borderWidth: 1,
    borderColor: theme.colors.border,
    borderRadius: theme.radius.md,
    padding: theme.spacing.md,
    backgroundColor: theme.colors.surface,
    gap: 8,
  },
  discussionTitle: {
    color: theme.colors.text,
    fontWeight: '700',
    fontSize: 15,
  },
  discussionText: {
    color: theme.colors.muted,
    fontSize: 13,
  },
  empty: {
    marginTop: 24,
    textAlign: 'center',
    color: theme.colors.muted,
  },
});
