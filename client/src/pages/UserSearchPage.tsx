import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useState } from 'react';
import { Alert, StyleSheet, Text, View } from 'react-native';
import { AppHeader } from '../components/AppHeader';
import { FormInput } from '../components/FormInput';
import { PrimaryButton } from '../components/PrimaryButton';
import { ProfileAvatar } from '../components/ProfileAvatar';
import { ScreenContainer } from '../components/ScreenContainer';
import { theme } from '../constants/theme';
import { searchUsersByEmailApi, sendFriendRequestApi } from '../api/api';
import { RootStackParamList } from '../navigation/types';
import { useAppSelector } from '../redux/hooks';
import { User } from '../types/models';

type Props = NativeStackScreenProps<RootStackParamList, 'UserSearch'>;

export const UserSearchPage = ({ navigation }: Props) => {
  const meId = useAppSelector((state) => state.auth.user?.id ?? '');
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [requestedIds, setRequestedIds] = useState<Record<string, boolean>>({});

  const runSearch = async () => {
    const trimmedQuery = query.trim();
    if (!trimmedQuery) {
      setResults([]);
      return;
    }

    setLoading(true);
    try {
      const users = await searchUsersByEmailApi(trimmedQuery);
      setResults(users.filter((user) => user.id !== meId));
    } catch (err) {
      Alert.alert('Ошибка', err instanceof Error ? err.message : 'Не удалось выполнить поиск');
    } finally {
      setLoading(false);
    }
  };

  const onAddFriend = async (userId: string) => {
    try {
      await sendFriendRequestApi(userId);
      setRequestedIds((current) => ({ ...current, [userId]: true }));
      Alert.alert('Готово', 'Запрос на добавление отправлен');
    } catch (err) {
      Alert.alert('Ошибка', err instanceof Error ? err.message : 'Не удалось отправить запрос');
    }
  };

  return (
    <ScreenContainer>
      <AppHeader title="Поиск пользователей" onActionPress={() => navigation.goBack()} actionIcon="arrow-back" />

      <View style={styles.card}>
        <FormInput
          label="Email пользователя"
          value={query}
          onChangeText={setQuery}
          autoCapitalize="none"
          keyboardType="email-address"
          placeholder="example@mail.com"
        />
        <PrimaryButton title="Найти" onPress={runSearch} loading={loading} />
      </View>

      <Text style={styles.sectionTitle}>Результаты</Text>
      {results.length === 0 ? <Text style={styles.empty}>Пока ничего не найдено</Text> : null}

      <View style={styles.results}>
        {results.map((user) => {
          const alreadyRequested = Boolean(requestedIds[user.id]);
          const isSelf = user.id === meId;

          return (
            <View key={user.id} style={styles.resultCard}>
              <View style={styles.row}>
                <ProfileAvatar uri={user.avatarUri || undefined} size={60} />
                <View style={styles.meta}>
                  <Text style={styles.name}>{user.name}</Text>
                  <Text style={styles.email}>{user.email}</Text>
                  <Text style={styles.city}>{user.city}</Text>
                </View>
              </View>

              <View style={styles.actions}>
                <PrimaryButton title="Открыть профиль" type="outline" onPress={() => navigation.navigate('Profile', { userId: user.id })} />
                {!isSelf ? (
                  <PrimaryButton
                    title={alreadyRequested ? 'Запрос отправлен' : 'В друзья'}
                    onPress={() => onAddFriend(user.id)}
                    disabled={alreadyRequested}
                  />
                ) : null}
              </View>
            </View>
          );
        })}
      </View>
    </ScreenContainer>
  );
};

const styles = StyleSheet.create({
  card: {
    borderRadius: theme.radius.lg,
    borderWidth: 1,
    borderColor: theme.colors.border,
    backgroundColor: theme.colors.surface,
    padding: theme.spacing.md,
    gap: theme.spacing.sm,
  },
  sectionTitle: {
    color: theme.colors.text,
    fontSize: 18,
    fontWeight: '700',
    marginTop: 2,
  },
  empty: {
    color: theme.colors.muted,
    fontSize: 13,
  },
  results: {
    gap: theme.spacing.sm,
  },
  resultCard: {
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
    gap: 2,
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
  city: {
    color: theme.colors.muted,
    fontSize: 13,
  },
  actions: {
    gap: theme.spacing.xs,
  },
});