import { Layout, Menu, Button, Typography, Space, Dropdown } from 'antd';
import {
  UserOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  TeamOutlined,
  MedicineBoxOutlined,
  BankOutlined,
  DollarOutlined,
  TagsOutlined,
  SettingOutlined,
  FileTextOutlined,
  AppstoreOutlined,
  InboxOutlined,
  SafetyCertificateOutlined,
  IdcardOutlined,
  KeyOutlined,
  AuditOutlined,
} from '@ant-design/icons';
import { useState, useMemo, type ReactNode } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../auth/useAuth';

const { Header, Sider, Content } = Layout;

type MenuDef = {
  key: string;
  icon: ReactNode;
  label: string;
  permission: string;
};

type AppShellProps = {
  children: ReactNode;
};

/** قالب اصلی اپ با سایدبار و هدر */
export function AppShell({ children }: AppShellProps) {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout, hasPermission } = useAuth();

  const allMenuItems: MenuDef[] = [
    { key: '/users', icon: <TeamOutlined />, label: 'کاربران', permission: 'users.read' },
    { key: '/roles', icon: <SafetyCertificateOutlined />, label: 'نقش‌ها', permission: 'roles.read' },
    { key: '/reception', icon: <MedicineBoxOutlined />, label: 'پذیرش', permission: 'reception.read' },
    { key: '/patients', icon: <IdcardOutlined />, label: 'بیماران', permission: 'patient.read' },
    { key: '/organization', icon: <BankOutlined />, label: 'سازمان', permission: 'organization.read' },
    {
      key: '/organization-packages',
      icon: <InboxOutlined />,
      label: 'بسته‌های تعرفه',
      permission: 'organization_packages.read',
    },
    { key: '/services', icon: <AppstoreOutlined />, label: 'خدمات', permission: 'services.read' },
    { key: '/fund', icon: <DollarOutlined />, label: 'صندوق', permission: 'fund.read' },
    { key: '/bank-accounts', icon: <BankOutlined />, label: 'حساب‌های بانکی', permission: 'bank_account.read' },
    { key: '/special-codes', icon: <KeyOutlined />, label: 'کد خاص', permission: 'special_code.read' },
    { key: '/regulations', icon: <AuditOutlined />, label: 'ضوابط', permission: 'regulation.read' },
    { key: '/tariff', icon: <TagsOutlined />, label: 'تعرفه', permission: 'tariff.read' },
    { key: '/logs', icon: <FileTextOutlined />, label: 'لاگ‌ها', permission: 'logs.read' },

    { key: '/settings', icon: <SettingOutlined />, label: 'تنظیمات', permission: 'security.settings' },
  ];

  const menuItems = useMemo(
    () =>
      allMenuItems
        .filter((item) => hasPermission(item.permission))
        .map((item) => ({
          key: item.key,
          icon: item.icon,
          label: item.label,
        })),
    [hasPermission],
  );

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} trigger={null} theme="light">
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontWeight: 'bold',
          }}
        >
          {collapsed ? 'TP' : 'طب پرداز'}
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />
          <Dropdown
            menu={{
              items: [
                {
                  key: 'profile',
                  icon: <UserOutlined />,
                  label: 'پروفایل',
                  onClick: () => navigate('/profile'),
                },
                {
                  key: 'logout',
                  icon: <LogoutOutlined />,
                  label: 'خروج',
                  onClick: () => logout(),
                },
              ],
            }}
          >
            <Space style={{ cursor: 'pointer' }}>
              <UserOutlined />
              <Typography.Text>
                {user?.name} {user?.family} ({user?.roleName})
              </Typography.Text>
            </Space>
          </Dropdown>
        </Header>
        <Content style={{ margin: 16, padding: 16, background: '#fff', borderRadius: 8 }}>
          {children}
        </Content>
      </Layout>
    </Layout>
  );
}
