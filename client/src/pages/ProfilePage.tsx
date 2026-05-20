import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useEffect, useState } from 'react';
import { Alert, StyleSheet, Text, View } from 'react-native';
import { AppHeader } from '../components/AppHeader';
import { ScreenContainer } from '../components/ScreenContainer';
import { RootStackParamList } from '../navigation/types';
import { fetchUserApi, fetchUserAvatarApi, sendFriendRequestApi } from '../api/api';
import { useAppSelector } from '../redux/hooks';
import { ProfileAvatar } from '../components/ProfileAvatar';
import { PrimaryButton } from '../components/PrimaryButton';

type Props = NativeStackScreenProps<RootStackParamList, 'Profile'>;

export const ProfilePage = ({ route, navigation }: Props) => {
  const userId = route.params.userId;
  const meId = useAppSelector((s) => s.auth.user?.id ?? '');
  const [user, setUser] = useState<any | null>(null);
  const [avatarUri, setAvatarUri] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      setLoading(true);
      try {
        const data = await fetchUserApi(userId);
        const avatar = await fetchUserAvatarApi(userId).catch(() => '');
        if (mounted) {
          setUser(data);
          setAvatarUri(avatar);
        }
      } catch (err) {
        Alert.alert('Ошибка', err instanceof Error ? err.message : 'Не удалось загрузить профиль');
      } finally {
        if (mounted) setLoading(false);
      }
    };
    void load();
    return () => {
      mounted = false;
    };
  }, [userId]);

  const onAddFriend = async () => {
    try {
      await sendFriendRequestApi(userId);
      Alert.alert('Готово', 'Запрос на добавление отправлен');
    } catch (err) {
      Alert.alert('Ошибка', err instanceof Error ? err.message : 'Не удалось отправить запрос');
    }
  };

  return (
    <ScreenContainer>
      <AppHeader title="Профиль пользователя" onActionPress={() => navigation.goBack()} actionIcon="arrow-back" />
      {user ? (
        <View style={styles.container}>
          <ProfileAvatar uri={avatarUri || undefined} size={96} />
          <Text style={styles.name}>{user.name}</Text>
          <Text style={styles.meta}>{user.city}</Text>
          {user.about ? <Text style={styles.about}>{user.about}</Text> : null}
          {meId !== userId ? <PrimaryButton title="Добавить в друзья" onPress={onAddFriend} /> : null}
        </View>
      ) : (
        <Text style={styles.empty}>Профиль не найден</Text>
      )}
    </ScreenContainer>
  );
};

const styles = StyleSheet.create({
  container: {
    alignItems: 'center',
    gap: 8,
    padding: 16,
  },
  name: {
    fontSize: 18,
    fontWeight: '700',
  },
  meta: {
    color: '#666',
  },
  about: {
    marginTop: 8,
    color: '#222',
    textAlign: 'center',
  },
  empty: {
    marginTop: 24,
    textAlign: 'center',
  },
});
