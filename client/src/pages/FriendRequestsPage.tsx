import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useEffect, useState } from 'react';
import { Alert, RefreshControl, ScrollView, StyleSheet, Text, View } from 'react-native';
import { fetchIncomingFriendRequestsApi, fetchUserAvatarApi, acceptFriendRequestApi, rejectFriendRequestApi } from '../api/api';
import { AppHeader } from '../components/AppHeader';
import { PrimaryButton } from '../components/PrimaryButton';
import { ProfileAvatar } from '../components/ProfileAvatar';
import { ScreenContainer } from '../components/ScreenContainer';
import { theme } from '../constants/theme';
import { RootStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<RootStackParamList, 'FriendRequests'>;

type IncomingRequestViewModel = {
  requestId: number;
  userId: string;
  name: string;
  avatarUri: string;
};

export const FriendRequestsPage = ({ navigation }: Props) => {
  const [requests, setRequests] = useState<IncomingRequestViewModel[]>([]);
  const [loading, setLoading] = useState(false);
  const [actionInProgress, setActionInProgress] = useState<Record<number, boolean>>({});

  const loadRequests = async () => {
    setLoading(true);
    try {
      const items = await fetchIncomingFriendRequestsApi();
      const mapped = await Promise.all(
        items.map(async (item) => {
          let avatarUri = '';
          try {
            avatarUri = await fetchUserAvatarApi(String(item.user.id));
          } catch {
            avatarUri = '';
          }
          return {
            requestId: item.request_id,
            userId: String(item.user.id),
            name: item.user.name,
            avatarUri,
          };
        })
      );
      setRequests(mapped);
    } catch (err) {
      Alert.alert('Ошибка', err instanceof Error ? err.message : 'Не удалось загрузить запросы');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadRequests();
  }, []);

  const updateRequestList = (requestId: number) => {
    setRequests((current) => current.filter((request) => request.requestId !== requestId));
  };

  const handleAccept = async (requestId: number) => {
    setActionInProgress((current) => ({ ...current, [requestId]: true }));
    try {
      await acceptFriendRequestApi(requestId);
      updateRequestList(requestId);
      Alert.alert('Готово', 'Запрос принят');
    } catch (err) {
      Alert.alert('Ошибка', err instanceof Error ? err.message : 'Не удалось принять запрос');
    } finally {
      setActionInProgress((current) => ({ ...current, [requestId]: false }));
    }
  };

  const handleReject = async (requestId: number) => {
    setActionInProgress((current) => ({ ...current, [requestId]: true }));
    try {
      await rejectFriendRequestApi(requestId);
      updateRequestList(requestId);
      Alert.alert('Готово', 'Запрос отклонён');
    } catch (err) {
      Alert.alert('Ошибка', err instanceof Error ? err.message : 'Не удалось отклонить запрос');
    } finally {
      setActionInProgress((current) => ({ ...current, [requestId]: false }));
    }
  };

  return (
    <ScreenContainer>
      <AppHeader title="Входящие запросы" onActionPress={() => navigation.goBack()} actionIcon="arrow-back" />

      <ScrollView
        refreshControl={<RefreshControl refreshing={loading} onRefresh={loadRequests} tintColor={theme.colors.primary} />}
        contentContainerStyle={styles.content}
        showsVerticalScrollIndicator={false}
      >
        <Text style={styles.caption}>Здесь отображаются запросы на добавление в друзья, которые можно принять или отклонить.</Text>

        {requests.length === 0 ? <Text style={styles.empty}>Пока нет входящих запросов</Text> : null}

        <View style={styles.list}>
          {requests.map((request) => (
            <View key={request.requestId} style={styles.card}>
              <View style={styles.row}>
                <ProfileAvatar uri={request.avatarUri || undefined} size={60} />
                <View style={styles.meta}>
                  <Text style={styles.name}>{request.name}</Text>
                  <Text style={styles.email}>ID: {request.userId}</Text>
                </View>
              </View>

              <View style={styles.actions}>
                <PrimaryButton
                  title="Принять"
                  onPress={() => handleAccept(request.requestId)}
                  loading={Boolean(actionInProgress[request.requestId])}
                />
                <PrimaryButton
                  title="Отклонить"
                  type="outline"
                  onPress={() => handleReject(request.requestId)}
                  loading={Boolean(actionInProgress[request.requestId])}
                />
              </View>
            </View>
          ))}
        </View>
      </ScrollView>
    </ScreenContainer>
  );
};

const styles = StyleSheet.create({
  content: {
    gap: theme.spacing.md,
    paddingBottom: theme.spacing.lg,
  },
  caption: {
    color: theme.colors.muted,
    fontSize: 13,
    lineHeight: 18,
  },
  empty: {
    color: theme.colors.muted,
    fontSize: 14,
  },
  list: {
    gap: theme.spacing.sm,
  },
  card: {
    borderRadius: theme.radius.lg,
    borderWidth: 1,
    borderColor: theme.colors.border,
    backgroundColor: theme.colors.surface,
    padding: theme.spacing.md,
    gap: theme.spacing.sm,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: theme.spacing.sm,
  },
  meta: {
    flex: 1,
    gap: 4,
  },
  name: {
    color: theme.colors.text,
    fontSize: 16,
    fontWeight: '700',
  },
  email: {
    color: theme.colors.primary,
    fontSize: 13,
  },
  actions: {
    gap: theme.spacing.xs,
  },
});