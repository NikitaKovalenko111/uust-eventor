import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useEffect, useState } from 'react';
import { Alert, Image, StyleSheet, Text, View } from 'react-native';
import { AppHeader } from '../components/AppHeader';
import { FormInput } from '../components/FormInput';
import { PrimaryButton } from '../components/PrimaryButton';
import { ProfileAvatar } from '../components/ProfileAvatar';
import { ScreenContainer } from '../components/ScreenContainer';
import { theme } from '../constants/theme';
import { createEventCommentApi, fetchEventCommentsApi, fetchUserAvatarApi, fetchEventAttendeesApi } from '../api/api';
import { RootStackParamList } from '../navigation/types';
import { useAppDispatch, useAppSelector } from '../redux/hooks';
import { registerForEventRequest, deleteEventRequest, finishEventRequest } from '../redux/slices/eventsSlice';
import { EventComment } from '../types/models';

type Props = NativeStackScreenProps<RootStackParamList, 'EventDetails'>;

export const EventDetailsPage = ({ route, navigation }: Props) => {
  const dispatch = useAppDispatch();
  const eventItem = useAppSelector((state) => state.events.list.find((event) => event.id === route.params.eventId));
  const userId = useAppSelector((state) => state.auth.user?.id ?? '');
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const [comments, setComments] = useState<EventComment[]>([]);
  const [commentsLoading, setCommentsLoading] = useState(false);
  const [commentsError, setCommentsError] = useState('');
  const [commentText, setCommentText] = useState('');
  const [sendingComment, setSendingComment] = useState(false);
  const [commentAvatars, setCommentAvatars] = useState<Record<string, string>>({});
  const [attendeesList, setAttendeesList] = useState<any[]>([]);

  useEffect(() => {
    let isMounted = true;

    const loadComments = async () => {
      if (!eventItem) {
        return;
      }

      setCommentsLoading(true);
      setCommentsError('');

      try {
        const loadedComments = await fetchEventCommentsApi(eventItem.id);

        const uniqueAuthorIds = Array.from(new Set(loadedComments.map((item) => item.authorId))).filter(Boolean);
        const avatarEntries = await Promise.all(
          uniqueAuthorIds.map(async (authorId) => {
            try {
              const avatarUri = await fetchUserAvatarApi(authorId);
              return [authorId, avatarUri] as const;
            } catch {
              return [authorId, ''] as const;
            }
          })
        );

        if (isMounted) {
          setComments(loadedComments);
          setCommentAvatars(Object.fromEntries(avatarEntries));
        }
      } catch (error) {
        if (isMounted) {
          setCommentsError(error instanceof Error ? error.message : 'Не удалось загрузить комментарии');
        }
      } finally {
        if (isMounted) {
          setCommentsLoading(false);
        }
      }
    };

    void loadComments();

    return () => {
      isMounted = false;
    };
  }, [eventItem?.id]);

  const isRegistered = eventItem?.attendees?.includes(userId) ?? false;
  const isOwnEvent = isAuthenticated && eventItem?.creatorId === userId;

  useEffect(() => {
    let mounted = true;
    const loadAttendees = async () => {
      if (!eventItem) return;
      try {
        // only load attendees for own events
        const list = await fetchEventAttendeesApi(eventItem.id);
        if (mounted && isAuthenticated && eventItem.creatorId === userId) setAttendeesList(list);
      } catch {
        if (mounted) setAttendeesList([]);
      }
    };
    void loadAttendees();
    return () => {
      mounted = false;
    };
  }, [eventItem?.id, isOwnEvent]);

  if (!eventItem) {
    return (
      <ScreenContainer>
        <AppHeader title="Мероприятие" onActionPress={() => navigation.goBack()} actionIcon="arrow-back" />
        <Text style={styles.empty}>Мероприятие не найдено</Text>
      </ScreenContainer>
    );
  }
  

  const submitComment = async () => {
    if (!isAuthenticated) {
      navigation.navigate('Auth');
      return;
    }

    const text = commentText.trim();
    if (!text) {
      Alert.alert('Комментарий пустой', 'Введите текст комментария');
      return;
    }

    setSendingComment(true);
    try {
      await createEventCommentApi(eventItem.id, { text });
      setCommentText('');

      const refreshedComments = await fetchEventCommentsApi(eventItem.id);
      const uniqueAuthorIds = Array.from(new Set(refreshedComments.map((item) => item.authorId))).filter(Boolean);
      const avatarEntries = await Promise.all(
        uniqueAuthorIds.map(async (authorId) => {
          try {
            const avatarUri = await fetchUserAvatarApi(authorId);
            return [authorId, avatarUri] as const;
          } catch {
            return [authorId, ''] as const;
          }
        })
      );

      setComments(refreshedComments);
      setCommentAvatars(Object.fromEntries(avatarEntries));
    } catch (error) {
      Alert.alert('Ошибка', error instanceof Error ? error.message : 'Не удалось отправить комментарий');
    } finally {
      setSendingComment(false);
    }
  };

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
      {eventItem.friendsCount ? <Text style={styles.meta}>{eventItem.friendsCount} ваших друзей зарегистрированы на мероприятие</Text> : null}

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
          <View style={{ marginTop: 8 }}>
            <Text style={{ fontWeight: '700', marginBottom: 6 }}>Зарегистрированные пользователи</Text>
            {attendeesList.map((a) => (
              <View key={a.id} style={{ flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 6 }}>
                <ProfileAvatar uri={a.avatarId || undefined} size={36} />
                <Text style={{ color: theme.colors.text }} onPress={() => navigation.navigate('Profile', { userId: a.userId })}>{a.name}</Text>
              </View>
            ))}
          </View>
        </View>
      ) : null}

      <View style={styles.discussionBlock}>
        <Text style={styles.discussionTitle}>Обсуждение мероприятия</Text>

        {isAuthenticated ? (
          <View style={styles.composer}>
            <FormInput
              label="Ваш комментарий"
              value={commentText}
              onChangeText={setCommentText}
              placeholder="Напишите, что думаете о мероприятии"
              multiline
              numberOfLines={4}
              style={styles.commentInput}
            />
            <PrimaryButton title="Отправить комментарий" onPress={submitComment} loading={sendingComment} />
          </View>
        ) : (
          <PrimaryButton title="Войти, чтобы оставить комментарий" type="outline" onPress={() => navigation.navigate('Auth')} />
        )}

        {commentsLoading ? <Text style={styles.commentsHint}>Загрузка комментариев...</Text> : null}
        {commentsError ? <Text style={styles.commentsError}>{commentsError}</Text> : null}
        {!commentsLoading && comments.length === 0 ? <Text style={styles.commentsHint}>Пока нет комментариев. Будьте первым.</Text> : null}

        {comments.map((comment) => (
          <View key={comment.id} style={styles.commentCard}>
            <ProfileAvatar uri={commentAvatars[comment.authorId] || undefined} size={42} />
            <View style={styles.commentBody}>
              <View style={styles.commentHeader}>
                <Text style={styles.commentAuthor}>{comment.authorName}</Text>
                <Text style={styles.commentDate}>{new Date(comment.createdAt).toLocaleString('ru-RU', { dateStyle: 'medium', timeStyle: 'short' })}</Text>
              </View>
              <Text style={styles.commentText}>{comment.text}</Text>
            </View>
          </View>
        ))}
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
  composer: {
    gap: theme.spacing.sm,
  },
  commentInput: {
    minHeight: 110,
    height: 110,
    textAlignVertical: 'top',
    paddingTop: 12,
  },
  commentsHint: {
    color: theme.colors.muted,
    fontSize: 13,
  },
  commentsError: {
    color: theme.colors.danger,
    fontSize: 13,
  },
  commentCard: {
    flexDirection: 'row',
    gap: theme.spacing.sm,
    paddingTop: theme.spacing.sm,
    borderTopWidth: 1,
    borderTopColor: theme.colors.border,
  },
  commentBody: {
    flex: 1,
    gap: 4,
  },
  commentHeader: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    gap: 8,
  },
  commentAuthor: {
    color: theme.colors.text,
    fontSize: 14,
    fontWeight: '700',
  },
  commentDate: {
    color: theme.colors.muted,
    fontSize: 11,
  },
  commentText: {
    color: theme.colors.text,
    fontSize: 14,
    lineHeight: 20,
  },
  empty: {
    marginTop: 24,
    textAlign: 'center',
    color: theme.colors.muted,
  },
});
